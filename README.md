# Problem Statement

- Build an API that fetches the number of Covid-19 cases in each state and in India and persist it in MongoDB.
- Using this data, build an API that takes the user's GPS coordinates as input and returns the total number of Covid-19 cases in the user's state and in India(assume India specific coordinates only) and the last update time of data.
- [Bonus Task] Deploy the app to Heroku using the Add-ons for the DBs used.
- [Bonus Task] Add a redis caching layer which caches data for 30 mins.

# Solution

## Design

- Task 1
    - The first endpoint is simply called /api, which fetches the data using the public API available [here](https://api.rootnet.in/covid19-in/stats/latest) and then saves it to MongoDB

    - This endpoint can be called at any point and it will simply overwrite **that day's** data in the Database. This will create problems in case we have a multi region server deployment (In which case we will have to save date a normalized manner, say UTC), but the scope is limited to India

    - To solve the problem we just need 1 document in Mongo. However, I have stored it on a per-day basis to keep the option of exposing historical data open. To make that work, a per-hour/other granularity may have been better, but I discarded it as of now to make the design simpler.

## Assumptions

- We do not perform any data sanity on the value returned from the public API giving us the Covid metrics. This has a chance of data loss but we can't perform any real sanity without understanding their SLO/SLI etc. Assumption here is that their data will always be returned correctly. 