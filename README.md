# TIME TRAVEL

Use `getbydate` and `getbytimestamp` to get records valid at a particular time. Other methods create new records, update existing ones, or delete.

Input is validated.

API endpoints respond with HTTP statuses:

-200 (OK - get, delete, put)

-201 (Created - post)

-404 (Not Found)

-409 (Conflict - e.g. cannot update record)

-500 (Server error)


## API endpoints for getting records valid at {date}

```
/{type}/getbydate/{insuredId}/{date}
/{type}/getbytimestamp/{insuredId}/{timestamp}
```


Use "insured" for {type} to get complete data.

Can also use "employee" or "address" for current state of those facts.


## GetResourceById ("GET")

`/{type}/id/{id:[0-9]+}`

## Create ("POST") - requires body
`/{type}/new`

## Update ("PUT") - requires body
`/{type}/update`

Adds new record for "employee" or "address" that reflects the change. Will reject if "employee" or "address" does not exist or if no change from the last update.

## Delete ("DELETE")

`/{type}/delete/{id:[0-9]+}`

Permanently deletes record (insured, employee, or insured address) and all of its history.

## ~TIME TRAVEL~

`getbydate` and `getbytimestamp` get records valid at `date` or `timestamp`

## GetResourceByDate ("GET")

`/{type}/getbydate/{insuredId}/{date}`

Gets records valid at end-of-day in system timezone

## GetResourceByTimestamp ("GET")

`/{type}/getbytimestamp/{insuredId}/{timestamp}`

Same as getbydate, but using integer timestamp for exact times

See API tests in api/api_test.go
