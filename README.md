# BSUIR Queue bot

Telegram bot, for managing queues in BSUIR. Uses IIS and Google Sheets API to manage the queues in tables according to the schedule in university

##Features
- Create and submit requests for labworks, managing them in a queue sorted by submittion time
- View the detailed queue in google sheets
- Cron-based clearing of google sheets and database from deleted labworks, resubmission of requests for unsuccessful labwork takes
- Admin system, which leverages acceptance of submission to the press of a single button

| Command       | Description                                                                                                    |
| --------------| ---------------------------------------------------------------------------------------------------------------|
| /help, /start | Getting basic info on bot and it's commands                                                                    |
| /assign       | Requesting admin privelligies on the group                                                                     |
| /join         | Sending request to group admin for joining group                                                               |
| /submit       | Submitting labwork request. Requires being part of the group                                                   |
| /revert       | Reverting to a previous state of request. For instance, choose subject -> choose date -> revert -> choose date |
| /add          | Creating a custom labwork for your group.                                                                      |
| /queue        | Send a queue for selected labwork as message                                                                   |                                             
| /table        | Sends a link to google sheet for your group                                                                    |

##Deploy

Run via docker
```sh
docker compose --run
```
Or, run locally, with setup of .env file
```sh
go build -o main.go main && ./main
```
Be careful, that on first setup it will ask you for OAuth2 permissions on Google Sheets, creating credentials.json and token.json files in /src