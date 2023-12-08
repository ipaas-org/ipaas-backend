## todo

how to create an application:

1. give git repo (and branch, defaults to default branch) //it sends a request to backend, returns if the repo exists and if the branch exists (it works for private repo as we use user token)
2. select a language (or autodetect)
3. select the port the webserver listens to or use env var $PORT
4. specify additional env vars (key value pairs)
5. specify a name of the application (for now it's not used but it will be used to create a dns record) //sends a request to backend to check if the name is available

backend operations:

1. check that the application name is available
2. insert the application in the db with status pending
3. push the application to the image builder queue
4. image builder gets the application from the queue
5. image builder update the application status to building
6. image builder pulls the repo
7. image builder builds the image
8. if the build fails and the fault was image builder then retry, if the fault was the user then update the status to failed and send to the queue a message that the build for application xxx failed
9. if the build is succesful then update the application with the image id to use and sends to the queue a message that the build was succesful
10. backend receives the message and updates the application status to starting
11. creates the container given all the info (image id, port, env vars, name)
12. starts the container
13. if starting the container fails retry for 3 times then update the status to failed
14. if the container is running then update the status to running

for future:

1. specify additional add ons like databases (databases, databases web views, ... (for now only db and db's web views))
2. add a volume to the application (specify name and mount path)
   2

## critical

- [x] let user create an application
- [x] add a service that sets a dns record for an application, for now we make a trasparent one that just returns the ip and port of the application
- [x] after you create an application return the state of the application and the application id
- [x] add endpoint to get all the applications of an user
- [x] let user delete an application
- [ ] let user forcefully restrt an application
- [ ] let the user update an application
  - visibility
  - name
  - listening port
  - envs
  - branch
  - code base (based on commit on that branch)
- [x] application status polling endpoint
- [x] create a database service
- [x] add a db system to visualize the database
- [ ] log reader system (will need pagination)
- [x] implement a crude version of warden
- [ ] email sender to user with stuff (db credentials, warden notifications, ...)

## important

- [x] rename oauth service to git provider
- [ ] add getGitRepos function in oauth service
- [ ] return the list of repos (name and branches) owned by the user (sorted by latest update)

## performance

- [ ] CreateApplicationFromApplicationIDandImageID calls the db too many times, should add a cache layer for the application and the users and update could be more granular so that it doesnt need to retrive the entire application to update it

## not foundamental

- [ ] paginate the list of repos
- [ ] add git service method that creates a channel for a repo that send info about updates to the repo
- [ ] repos shouldnt return a bool in updates and delete cause it's not really used
