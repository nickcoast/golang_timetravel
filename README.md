# The Temelpa Take-Home Assignment

Please create a __private__ version of this repo, complete the objectives, and once you
are finished, send a link to your repo to us.

# The Assignment

Part of what an insurance company needs to have in its backend is a 
record system. As an insurer, we need to keep an up-to-date record of each of our policy-holder's
data points that go into the calculation of their rate. When a policy-holder updates
their information, I.E. they change addresses, or add/remove new employees to their team
we will be notified and we must keep our records up to date.

The current version of the repo is an extremely simplified version of exactly that. `GET /api/v1/record/{id}`
will retrieve a record, which is just a json mapping strings to strings. and `POST /api/v1/record/{id}`
will either create a new record or modify an existing record. However, it isn't enough to
just keep a record of the current record state but we must maintain a reference to how the state
has changed to be in full compliance.

Say that the policy-holder buys their insurance on the start of the year, and then two months later
changes the address of their business but doesn't tell us about this change until 4 months after that.
Since we were technically held liable if there was a claim event, we need to charge the customer the
difference for the 4 months since they changed addresses. To do so accurately, we need to know the
version of the records that we knew about them at the two points of time: at the time when the change happened
and at the time when we were told of the change.

In this project, you'll make a simplified version of this system. At a high-level your goal is to do two things:

1. Change the storage backend to sqlite, and persist the data across turning off and on the server.
2. Add the time travel component so we can easily look up the state of each records at different timesteps.

The sections below outline these two objectives in more detail. You may use whatever libraries and tools
you like to achieve this even as far as building this in an entirely different language.

## Objective: Switch To Sqlite

The current implementation does not store the data. The data is lost once the server 
process is killed. You should change the code so that all changes are persisted on 
to sqlite.

Once you're done, the data should be persistent on to a sqlite file as the server 
is running. The server should tolerate restarting the process without data loss.

## Objective: Add Time Travel

This part is far more open-ended. You might need to make major changes across nearly
all files of the codebase. You'll be adding partial persistentence to the records. [Take a look 
at these slides on persistent data structures to learn what that means](https://www.ics.uci.edu/~eppstein/261/s21w9.pdf).

You should create a set of `/api/v2` endpoints that enable you to do run gets, creates, and updates. 
Updates should be able to add modifications on top of any pre-existing versions not just the latest. 
There should be a way to get a list of the different versions too. Finally, `/api/v1` 
should still work after these changes.

# Reccommendations

We expect you to work as if this task was a normal project at work. So please write
your code in a way that fits your intuitive notion of operating within best practices.
Additionally, you should at the very least have a different commmit for each individual objective, 
ideally more as you go through process of completing the take-home. Also we like
to see your thought process and fixes as you make changes. So don't be afraid of
committing code that you later edit. No need to squash those commits.

Many parts of the assignment is intentionally ambiguious. If you have a question, definitely
reach out. But for many of these ambiguiuties, we want to see how you independently make
software design decisions.


# Reference -- The Current API

There are only two API endpoints `GET /api/v1/records/{id}` and `POST /api/v1/records/{id}`, all ids must be positive integers.

### `GET /api/v1/records/{id}`

This endpoint will return the record if it exists.

```bash
> GET /api/v1/records/2323 HTTP/1.1

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
{"id":2323,"data":{"david":"hey","davidx":"hey"}}
```

```bash
> GET /api/v1/records/32 HTTP/1.1

< HTTP/1.1 400 Bad Request
< Content-Type: application/json; charset=utf-8
{"error":"record of id 32 does not exist"}
```

### `POST /api/v1/records/{id}`

This endpoint will create a record if a does not exists.
Otherwise it will update the record.

The payload is a json object mapping strings to strings
and nulls. Values that are null indicate that the
backend must delete that key of the record.

```bash
# Creating a record
> POST /api/v1/records/1 HTTP/1.1
{"hello":"world"}

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
{"id":1,"data":{"hello":"world"}}


# Updating that record
> POST /api/v1/records/1 HTTP/1.1
{"hello":"world 2","status":"ok"}

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
{"id":1,"data":{"hello":"world 2","status":"ok"}}


# Deleting a field
> POST /api/v1/records/1 HTTP/1.1
{"hello":null}

< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
{"id":1,"data":{"status":"ok"}}
```
