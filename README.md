[![Go Report Card](https://goreportcard.com/badge/PalAditya/GoTasks)](https://goreportcard.com/report/PalAditya/GoTasks) [![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/PalAditya/GoTasks/blob/master/LICENSE) [![codecov](https://codecov.io/gh/PalAditya/GoTasks/branch/master/graph/badge.svg?token=5HIWD8XAE5)](https://codecov.io/gh/PalAditya/GoTasks)

# Problem Statement

- Build an API that fetches the number of Covid-19 cases in each state and in India and persist it in MongoDB.
- Using this data, build an API that takes the user's GPS coordinates as input and returns the total number of Covid-19 cases in the user's state and in India(assume India specific coordinates only) and the last update time of data.
- [Bonus Task] Deploy the app to Heroku using the Add-ons for the DBs used.
- [Bonus Task] Add a redis caching layer which caches data for 30 mins.

# Solution

## Design

- **Task 1**
    - The first endpoint is simply called /api, which fetches the data using the public API available [here](https://api.rootnet.in/covid19-in/stats/latest) and then saves it to MongoDB

    - This endpoint can be called at any point and it will simply overwrite **that day's** data in the Database. This will create problems in case we have a multi region server deployment (In which case we will have to save date a normalized manner, say UTC), but the scope is limited to India

    - To solve the problem we just need **1 document** in Mongo. However, I have stored it on a per-day basis to keep the option of exposing historical data open. To make that work, a per-hour/other granularity may have been better, but I discarded it as of now to make the design simpler.

    - The API looks as follows (The Swagger docs can be viewed by using `swagger serve -F=swagger swagger.yaml`)

    ![Structure](https://user-images.githubusercontent.com/25523604/133878816-91cc0c1c-4388-4924-814e-1a2c3621f278.png) ![Response](https://user-images.githubusercontent.com/25523604/133879285-2810b539-2f07-4a59-b211-94dd15b13971.png)

- **Task 2**

    - The second endpoint is **/data/{lat}/{long}**. First, it uses the reverse geocoding API provided by locationIQ to find the user's state. Next it fetches the latest document available in Mongo (going by `recordDate`) and iterates through the states to find the user's state and returns the results. It also returns a 404 resposne if the state can't be mapped to any state in DB (location outside India)

    - Swagger representation is as below

    ![Structure](https://user-images.githubusercontent.com/25523604/133879176-c6148bf5-f804-44c0-a66f-cb7c7d0a4973.png)
    ![Response](https://user-images.githubusercontent.com/25523604/133879180-a1e68ba0-14a5-41d5-800d-8db137415c9f.png)

- **Task 3**
    - [App](https://cryptic-river-61900.herokuapp.com/api) has been deployed to Heroku. We did not use the Add-Ons for Heroku, but rather used Cloud Atlas MongoDB since the Heroku add-on was not under a free tier

- **Task 4**
    - Currently, the design on Redis is extremely simple. When the `/data/{lat}/{long}` endpoint is invoked, it still fetches the User's state from the reverse GeoCoding API and *then searches for that key in Redis*. So we have **(state, mongoDoc)** in Redis with a TTL of **30 minutes**
    - There are several problems with the above idea, mainly the fact that there will be an update contention. Let's say the Mongo Data is also updated every half an hour - then we might show upto an hour's old data via Redis!
    - A shorter update schedule for Mongo (say 5 minutes) ensures Redis will never lag too far behind.
    - With authentication and authorization, caching can be greatly enhanced by using a combination of username, latitude and longitude which can allow us to skip the reverse geocoding call as well.
    - We can also update Redis using the below alogorithm (without setting TTL of 30 minutes):
        - Always add time of key creation/update to redis
        - If time >= 30 minutes, fetch from Mongo
        - If record in Mongo = Record in Redis, do nothing and return that result to user
        - Else, update the time of creation in Redis and return
    - The above will lead to 'bursts' when all calls might concurrently bypass Redis, but it will probably be the most correct one (implemented, unused. See `func IsPresentInCacheWithTime(key string)` for details). However, we can't use it without understanding the update schedule of Mongo first. To see it in action, switch to commit `8d9ad39d75810ef792cf270898419588e965957e` (Play with the timing of key expiry in Redis, 0 at that commit)
    - We are also using Redis to fetch the id of the latest doc in Mongo via using the `recordDate` as key since that won't change and is useful for `/api` endpoint's response
    - Heroku has **not been** provisioned with the Redis add-on (because it won't accept my card :smile:)

- **Extra Features**
    - Middlewares used: **CORS** (We get nice logging and time taken from Heroku itself), **Prometheus**
    - Grafana dashboard for the endpoints (Not exported to heroku. To view, one will have to download the dependencies himself/herself. Maybe can be included as a Docker script, but was unable to complete it in time. The grafana binaries are available [here](https://grafana.com/grafana/download?platform=windows) and prometheus binaries are [here](https://prometheus.io/download/))
    - Dashboard View (Json present in src):
    ![Status](https://user-images.githubusercontent.com/25523604/133895142-ae09a8fc-a891-4f8c-8ec0-35917a79b6b8.png)
    ![Latencies](https://user-images.githubusercontent.com/25523604/133895182-cceb8868-4213-46c8-b095-c620f22781bd.png)
    - Unit tests have been added for the `apis` package with `> 50%` coverage. Started working on the DB interactions too but it will take a bit more effort to mock them out


## Assumptions

- We do not perform any data sanity on the value returned from the public API giving us the Covid metrics. This has a chance of data loss (updating mongo with lesser data compared to what we have now) but we can't perform any real sanity without understanding their SLO/SLI etc. Assumption here is that their data will always be returned correctly. 
- Context: We are passing around the same context everywhere with time limit of **10 seconds**! This can be catastrophic, especially since we never cancel the context (The cancel handler is discarded). However, it was done to keep the readability high. Absolutely can't be done in Production.
- No retry mechanism has been provided to automatically recover from errors. No burst protection has been provided either but that would again require an authentication/authorization system
- Redis layer never really discards the keys (except on TTL), but it will never go beyond a certain size due to this implementation itself :smile: