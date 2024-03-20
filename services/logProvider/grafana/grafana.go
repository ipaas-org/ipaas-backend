package grafana

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ipaas-org/ipaas-backend/model"
	logprovider "github.com/ipaas-org/ipaas-backend/services/logProvider"
	"github.com/tidwall/gjson"
)

var _ logprovider.LogProvider = new(GrafanaLogProvider)

type GrafanaLogProvider struct {
	serviceAccountToken string
	grafanaUrl          *url.URL
	lokiUid             string
}

func NewGrafanaLogProvider(ctx context.Context, serviceAccountToken, grafanaUrl string) (*GrafanaLogProvider, error) {
	parsedGrafanaUrl, err := url.Parse(grafanaUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse grafana url (%s): %w", grafanaUrl, err)
	}

	g := &GrafanaLogProvider{
		serviceAccountToken: serviceAccountToken,
		grafanaUrl:          parsedGrafanaUrl,
	}

	if err := g.getLokiUid(ctx); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *GrafanaLogProvider) doRequest(ctx context.Context, method, endpoint string, body *bytes.Buffer) (*http.Response, error) {
	var req *http.Request
	var err error
	url := g.grafanaUrl.JoinPath(endpoint).String()
	if body != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, body)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.serviceAccountToken))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func (g *GrafanaLogProvider) getLokiUid(ctx context.Context) error {
	//get loki datasource uid
	resp, err := g.doRequest(ctx, "GET", "api/datasources", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body := string(buf)
	lokiUid := ""
	gjson.Parse(body).ForEach(func(key, value gjson.Result) bool {
		lokiUid = value.Get("uid").String()
		return !(value.Get("type").String() == "loki")
	})
	if lokiUid == "" {
		return fmt.Errorf("loki datasource not found")
	}
	g.lokiUid = lokiUid
	return nil
}

func (g *GrafanaLogProvider) GetLogs(ctx context.Context, namespace string, app string, from string, to string) (*model.LogBlock, error) {
	toTime, err := g.parseGrafanaTime(to)
	if err != nil {
		return nil, fmt.Errorf("failed to parse to time: %s", err)
	}
	fromTime, err := g.parseGrafanaTime(from)
	if err != nil {
		return nil, fmt.Errorf("failed to parse from time: %s", err)
	}
	logBlock := &model.LogBlock{
		From:      fromTime,
		To:        toTime,
		Namespace: namespace,
		App:       app,
	}

	reqBody := g.generateBody(
		namespace,
		app,
		strconv.Itoa(int(fromTime.UnixMilli())),
		strconv.Itoa(int(toTime.UnixMilli())))
	resp, err := g.doRequest(ctx, "POST", "/api/ds/query", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := string(buf)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get logs from grafana: %s", resp.Status)
	}

	size := g.getLogFullSize(body)
	logBlock.TotalLogs = size

	logs := g.retriveLogContents(body)
	logBlock.ReturnedLogs = len(logs)
	logBlock.Content = logs
	if logBlock.ReturnedLogs > 0 {
		logBlock.LastTimestamp = logs[len(logs)-1].Timestamp
	}
	return logBlock, nil
}

func (g *GrafanaLogProvider) parseGrafanaTime(grafanaTime string) (time.Time, error) {
	if grafanaTime == "now" {
		return time.Now(), nil
	}
	if strings.HasPrefix(grafanaTime, "now-") {
		duration, err := time.ParseDuration(grafanaTime[4:])
		if err != nil {
			return time.Time{}, err
		}
		return time.Now().Add(-duration), nil
	}
	return time.Parse(time.RFC3339Nano, grafanaTime)
}

func (g *GrafanaLogProvider) getLogFullSize(body string) int {
	var size int
	for _, result := range gjson.Get(body, "results.ipaas.frames.0.schema.meta.stats").Array() {
		if result.Get("displayName").String() == "Summary: total lines processed" {
			size = int(result.Get("value").Int())
			break
		}
	}
	return size
}

func (g *GrafanaLogProvider) retriveLogContents(body string) []model.LogContent {
	blocks := gjson.Get(body, "results.ipaas.frames.0.data.values").Array()
	contents := blocks[2].Array()
	tsNano := blocks[3].Array()
	size := len(contents)
	logs := make([]model.LogContent, size)
	for i := range size {
		// logs[i].Fields = make(map[string]string)
		// for key, value := range fields[i].Map() {
		// 	logs[i].Fields[key] = value.String()
		// }
		logs[i].Content = contents[size-1-i].String()
		logs[i].Timestamp = time.Unix(0, tsNano[size-1-i].Int())
	}
	return logs
}

func (g *GrafanaLogProvider) generateBody(ns, app, from, to string) *bytes.Buffer {
	reqBody := fmt.Sprintf(`{"queries":[{"expr":"{namespace=~\"%s\", stream=~\".+\", app =~\"%s\"} |= \"\"","refId":"ipaas","datasource":{"type":"loki","uid":"%s"},"queryType":"range","maxLines":5000,"legendFormat":"","datasourceId":5,"intervalMs":1000,"maxDataPoints":1030}],"from":"%s","to":"%s"}`,
		ns, app, g.lokiUid, from, to)
	return bytes.NewBuffer([]byte(reqBody))
}
