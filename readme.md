# IPAAS

## Itis Paleocapa As A Service

Itis Paleocapa as a Service (abbreviated IPaaS) is a webapp dedicated to the students of the [I.T.I.S. Pietro Paleocapa](https://www.itispaleocapa.edu.it) of Bergamo.
Through an access guaranteed by [PaleoID](https://github.com/cristianlivella/paleoid-backend), the application allows users to host their web applications on the school server.

**Description:**

As anticipated, the program allows users to distribute their application on the school network servers, thus providing a useful and concrete tool for all the developers inside the institute.
The main difference compared to other competitors in the sector lies in the simplification of use for students. The only requirement is to have an email from the institution, without requiring a credit card to verify its authenticity.
Furthermore, IPaaS does not limit the number of applications that can be hosted by a single user, does not impose a maximum hour limit for hosted applications, and does not require payments or subscriptions of any kind.

### Requirements

- docker-compose
- docker: make sure you have sudo privileges on the docker group (check this out to know how to do so [docker post-installation on linux](https://docs.docker.com/engine/install/linux-postinstall/)), if you don't wanna do that tho then run `go build .` and run the binary as sudo
- required images (to install them run `docker pull <image name>`:
  - golang:1-alpine3.15
  - mysql:8.0.28-oracle
  - mariadb:10.8.2-rc-focal
  - mongo:5.0.6

### How to use

- Make sure to create a .env environent following the .env.example file
- run `$ docker-compose up --build -d`
- go run .

_**must create a .env file with the correct configurations**_

### Example

if you don't have a paleoID identity (an email that has as domain @itispaleocapa.it) but still would
like to test the application then do a post request to:
`/api/mock/create`
with body a raw json with such fields:
`{ "password":"aNicePassWord", "name": "aNiceName", "userID":"1234" } `

Make sure that userID is a number.
If an error is returned with body `User already exists` it means that another user has that userID so choose another one.

**This account will be closed after a day of being created as it's a test user used to test to behaviour of the application.**

If you prefer there is a web page at /mock where you can create a mock user using a gui.

you can use this repo [vano2903/testing](https://github.com/Vano2903/testing/) as a testing webserver
