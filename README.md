# TIME TRAVEL

Use getbydate and getbytimestamp to get records valid at a particular time

API endpoints for getting records valid at {date}
/{type}/getbydate/{insuredId}/{date}
/{type}/getbytimestamp/{insuredId}/{date}

Use "insured" for {type} to get complete data.
Can also use "employee" or "address" for current state of those facts.


GetResourceById ("GET")
"/{type}/id/{id:[0-9]+}"

Create ("POST")
"/{type}/new"

Update ("PUT")
"/{type}/update"

Delete ("DELETE")
"/{type}/delete/{id:[0-9]+}"
Permanently deletes record (insured, employee, or insured address)

!!!TIME TRAVEL!!! - use getbydate and getbytimestamp to get records valid at a particular time

GetResourceByDate ("GET")
"/{type}/getbydate/{insuredId}/{date}"
// Gets records valid at end-of-day in system timezone

GetResourceByTimestamp ("GET")
"/{type}/getbytimestamp/{insuredId}/{date}"
same as getbydate, but using integer timestamp for exact times
	

Input is validated.
API endpoints respond with HTTP statuses:
200 (OK - get, delete, put)
201 (Created - post)
404 (Not Found)
409 (Conflict - e.g. cannot update record)
500 (Server error)

They also try to return informative messages.



See API tests in api/api_test.go

