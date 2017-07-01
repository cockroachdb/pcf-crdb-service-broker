# CockroachDB service broker for Pivotal Cloud Foundry

## Introduction

The CRDB service broker allows binding applications to CockroachDB instances,
providing "Level 2" integration (see [pivotal
docs](https://docs.pivotal.io/tiledev/brokered.html)).

A service broker is configured to expose one or more *service plans*. Each
service plan is associated to a CockroachDB cluster or instance.

From service plans we can instantiate *service instances*. For each service
instance, a database (namespace) is created on a CockroachDB instance.

Once we have a service set up, we can *bind* it to applications. Each binding
creates a unique user/password for that application and grants permissions on
the service database. The credentials are passed via a `postgres://` URI;
applications which support Postgres may work without modifications (assuming
they don't use postgres-specific syntax that CockroachDB doesn't support).

Note that when we want different apps to be separated (different namespaces),
they should use different service instances.

## How to use

### Prerequisites

- Install the [CF CLI](https://docs.cloudfoundry.org/cf-cli/install-go-cli.html).
- [Login](https://docs.cloudfoundry.org/cf-cli/getting-started.html#login) to a CF instance.
- Set up one or more CockroachDB instances. These don't need to live on CF.
  Note: for things to work with the sample `spring-music` app, a recent version
  of CockroachDB which supports
  [int4](https://github.com/cockroachdb/cockroach/pull/16720) must be used.


### Installing the service broker

The service broker can either be pushed as an app, or deployed from a tile.

#### Using `cf push`

- Edit `manifest.yml` and edit the `PRECONFIGURED_PLANS` variable. This variable
  is a list of service plans; each plan is configured to connect to a
  CockroachDB instance (this can be a single node, or a load balancer on top of
  a cluster).

- Push the service broker "app":
  ```
  cf push
  ```
  You should see the service broker app and hostname under `cf apps`.

- Create the service broker:
  ```
  cf create-service-broker crdb-service-broker user pass https://<hostname>
  ```
  The user and password are `SECURITY_USER_NAME` and `SECURITY_USER_PASSWORD`
  in `manifest.yml`; the hostname is the one shown by `cf apps`.

The `cf service-brokers` command should show `crdb-service-broker`.


#### Using the tile

The service broker is available as a CF tile. You can build the tile using the
[Tile Generator](https://docs.pivotal.io/tiledev/tile-generator.html) by running
`build.sh`; then upload the `release/*.pivotal` file to CF (using the ops
manager). Then you can install the tile; there is a configuration form for
specifying details for the service plans.

Note that every build bumps the tile version. CF will barf if it sees two files
with the same version that are different (even if the old one was uninstalled),
so it's best to not circumvent the version change.

After the installation, the `cf service-brokers` command should show `crdb-service-broker`.

### Using the CockroachDB service

Make sure access to the service is enabled via `cf service-access`. To enable,
use `cf enable-service-access cockroachdb`. Then `cf marketplace` should list
the `cockroachdb` service with whatever plans the broker was pre-configured.

#### Create a service instance

Create a service for a specific plan:
```
cf create-service cockroachdb default crdb-service-1
```

If we log into the database, we can see that a new database has been created.

```sql
root@52.170.84.221:26257/> SHOW DATABASES;
+-------------------------------------+
|              Database               |
+-------------------------------------+
| cf_gbccnddiddnnhfdolliklaolojgmceif |
| crdb_internal                       |
| information_schema                  |
| lol                                 |
| pg_catalog                          |
| system                              |
+-------------------------------------+
(6 rows)
```

#### Bind service instance to app

We can test things with a sample app called `spring-music`, which just exposes a
web UI to insert/remove music album information. The application supports a
postgres backend. To install it, see the [`spring-music-only-war` repo](https://github.com/svennela/spring-music-only-war).

Once `spring-music` is pushed, we can bind it to our new service:
```
cf bind-service spring-music crdb-service-1
```

This operation created a new user with access to the database associated with
`crdb-service-1`:
```sql
root@52.170.84.221:26257/> SHOW USERS;
+----------------------------------+
|             username             |
+----------------------------------+
| dkbafmddnfmgofgghiflhkcbnkdleknp |
| test                             |
+----------------------------------+
(2 rows)
root@52.170.84.221:26257/> SHOW GRANTS ON DATABASE cf_gbccnddiddnnhfdolliklaolojgmceif;
+-------------------------------------+----------------------------------+------------+
|              Database               |               User               | Privileges |
+-------------------------------------+----------------------------------+------------+
| cf_gbccnddiddnnhfdolliklaolojgmceif | dkbafmddnfmgofgghiflhkcbnkdleknp | ALL        |
| cf_gbccnddiddnnhfdolliklaolojgmceif | root                             | ALL        |
+-------------------------------------+----------------------------------+------------+
```

We can take a look at the environment for spring-music, and notice it includes a
URI to our database:
```
cf env spring-music
...
     "uri": "postgres://dkbafmddnfmgofgghiflhkcbnkdleknp:3SJ2dD3K6v1PfHk1@52.170.84.221:26257/cf_gbccnddiddnnhfdolliklaolojgmceif?sslmode=disable",

...
```
Restart the app (`cf restart spring-music`) and it should now be using our database:
```sql
root@52.170.84.221:26257/> SHOW TABLES FROM cf_gbccnddiddnnhfdolliklaolojgmceif;
+-------+
| Table |
+-------+
| album |
+-------+
(1 row)
root@52.170.84.221:26257/> SELECT * FROM cf_gbccnddiddnnhfdolliklaolojgmceif.album;
+--------------------------------------+---------+---------------------------+-------+-------------+----------------------------+------------+
|                  id                  | albumid |          artist           | genre | releaseyear |           title            | trackcount |
+--------------------------------------+---------+---------------------------+-------+-------------+----------------------------+------------+
| 24c8fb84-9908-488a-aeb8-ab3ccae6df83 | NULL    | Michael Jackson           | Pop   |        1982 | Thriller                   |          0 |
| 27ce02b2-f6d6-4679-bc61-4c059d015088 | NULL    | Jimi Hendrix Experience   | Rock  |        1967 | Are You Experienced?       |          0 |
| 292de30a-a1e6-4585-92af-f09d699ccb20 | NULL    | Fleetwood Mac             | Rock  |        1977 | Rumours                    |          0 |
| 2e43897e-ff0c-4c77-975c-c8f3da6ae8fc | NULL    | The Beatles               | Rock  |        1969 | Abbey Road                 |          0 |
| 31f90083-9dac-405b-af04-e8186210294b | NULL    | The Rolling Stones        | Rock  |        1969 | Let it Bleed               |          0 |
| 357512b3-480b-482c-8dd5-73db3158f548 | NULL    | Led Zeppelin              | Rock  |        1969 | Led Zeppelin               |          0 |
| 41fd0674-44fe-47d8-b4ae-6bc80f8431cb | NULL    | Stevie Ray Vaughan        | Blues |        1983 | Texas Flood                |          0 |
| 4526d75e-aa48-4dd9-ac92-a6e8683818b2 | NULL    | The Eagles                | Rock  |        1976 | Hotel California           |          0 |
| 52e72d98-b535-435f-b688-1d7fe58985c6 | NULL    | The Clash                 | Rock  |        1980 | London Calling             |          0 |
| 5a94e4e8-d656-40b3-8699-183baacdbdfb | NULL    | Robert Johnson            | Blues |        1961 | King of the Delta Blues    |          0 |
| 609b5b41-4759-4ae3-b699-0c8356a9a409 | NULL    | The Rolling Stones        | Rock  |        1972 | Exile on Main Street       |          0 |
| 71b5ba16-4509-4966-b428-ebdb8a0706fb | NULL    | The Ramones               | Rock  |        1976 | The Ramones                |          0 |
| 7c02d294-81b3-4fae-bb65-accd0a9de511 | NULL    | Boston                    | Rock  |        1978 | Don't Look Back            |          0 |
| 7dfb002c-f8ee-477a-bf5d-eb40b1689107 | NULL    | The Fabulous Thunderbirds | Blues |        1979 | Rock With Me               |          0 |
| 7e7d5ecb-419c-43cf-b004-cc50c2068ff7 | NULL    | Queen                     | Rock  |        1975 | A Night At The Opera       |          0 |
| 875ca773-7b5f-4a56-bae5-cf762749c7c0 | NULL    | Albert King               | Blues |        1967 | Born Under A Bad Sign      |          0 |
| 91e7751d-705b-4f1a-8d08-7910cfd294bd | NULL    | Nirvana                   | Rock  |        1991 | Nevermind                  |          0 |
| 93732bce-994c-45e4-b431-a36d7e92773d | NULL    | Bruce Springsteen         | Rock  |        1975 | Born to Run                |          0 |
| 98f80b7c-e28b-4aec-bfdc-f447e687d6c1 | NULL    | The Beach Boys            | Rock  |        1966 | Pet Sounds                 |          0 |
| 9d6c043a-90c1-4764-aa6e-88ec42275617 | NULL    | Marvin Gaye               | Rock  |        1971 | What's Going On            |          0 |
| 9f394353-d0d6-413f-a392-3a9a8b23e625 | NULL    | U2                        | Rock  |        1991 | Achtung Baby               |          0 |
| a7aa43fb-c712-4204-8705-90b7ec27c426 | NULL    | Police                    | Rock  |        1983 | Synchronicity              |          0 |
| a7ab9a9c-fca6-4f86-8023-d8ae6b19ae08 | NULL    | The Beatles               | Rock  |        1965 | Rubber Soul                |          0 |
| b8388fef-3df1-48b2-a9fe-2ceca6cfa5b0 | NULL    | U2                        | Rock  |        1987 | The Joshua Tree            |          0 |
| c72cdab7-d333-4502-83ac-1a95923d2dad | NULL    | Stevie Ray Vaughan        | Blues |        1984 | Couldn't Stand The Weather |          0 |
| d5fac975-ff76-42e9-aee4-57ed23032101 | NULL    | Elvis Presley             | Rock  |        1976 | Sun Sessions               |          0 |
| d7621a16-1fca-4878-935f-bc12ea20e72a | NULL    | BB King                   | Blues |        1956 | Singin' The Blues          |          0 |
| f03ca1a7-9cf3-4294-b447-6b237d9df058 | NULL    | Led Zeppelin              | Rock  |        1971 | IV                         |          0 |
| f0a93b83-9029-4102-a208-b808a0628bc3 | NULL    | Muddy Waters              | Blues |        1964 | Folk Singer                |          0 |
+--------------------------------------+---------+---------------------------+-------+-------------+----------------------------+------------+
(29 rows)
```
