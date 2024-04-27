# every log

a logging service you can use anywhere.

## idea

be able to load some environment variables, import a package and then have your logs visible in a pretty frontend.

this is something i'm leaving in public as i work on it, but may take private at some stage. as such the [license](LICENSE) is more restrictive than my other public projects.

## workflow

- load environment variables for EVERYLOG_USER, EVERYLOG_PROJECT and optionally EVERYLOG_TEAM.
- import the everylog sdk for your language.
- insert your logs into your project.
- view your logs in the frontend.

## basic example

```python
import os
from everylog import EveryLog

# get environment variable
EVERYLOG_API_KEY = os.environ.get("EVERYLOG_API_KEY")
# create logger
logger = EveryLog(key=EVERYLOG_API_KEY)
# basic global project info log
logger.info("hello world!")
# alternatively, specify a process
logger.info("log for the main process", process="main")
# you could also create a logger per function and make a process for that function.
# the idea would be to construct this object at the top level of any process and pass it around.
def hello_world():
    logger = EveryLog(key=EVERYLOG_API_KEY, process="hello_world")
    logger.info("hello world from this process!")
```

## design

### database considerations

- PII data should be isolated so that permissions for certain consumers can be limited.
- Locations should have their own discrete model so they can be easily reused and are less tied to individual users and orgs.
- A User should always have to create a project, and can then optionally assign ownership of it to an Org they belong to.
- When a User creates a Project there should be a transaction that also creates a PermittedProject linking the User to the Project.
- Multiple Org should be able to collaborate on a Project.
- When an Org is linked to a Project using ProjectOrg, admins can assign their users defined by UserOrg to the Project, creating a PermittedProject
- A user will generate a new api key for each project they work on. As such the Api Key will be related to PermittedProject to ensure a user can have max 1 API key per project.
- Expiration time is omitted from the AuthorizationToken model so that it can be derived from the created_at on the server.

### server apis

All requests should require an Accept header of "application/json" so they can be expanded to return html later.

The endpoints are as follows. Endpoint parameters assume user_id and an authorization_token are passed in the request headers.

They are also ordered in the way a first time user would set themselves up in the CLI, and the way the onboarding flow should work.

#### No auth

- [x] POST /user (email, first_name, optional last_name, optional mobile_number) -> user_id
      From here, every endpoint should include user_id in the request header

#### Authentication auth (email/password currently)

- [x] POST /authenticate(email, password) -> authorization_token
- [x] POST /authorize(authorization_token) -> authorization_token (internal, for handling authorization token in header)

#### User auth (token)

Note any of these endpoints could return an unauthorized if the token has expired.

- [ ] POST /project (user_id, name, optional description) -> project_id (New Project)
- [ ] POST /project/{project_id}/key (email, password) -> api_key (Get API key for project)
- [ ] POST /log (level_id, project_id, message, optional process_id, optional traceback) (Create Log)
- [ ] GET /log (optional projectId, optional level_id, optional process_id, optional org_id, optional from_datetime, optional to_datetime) -> Array<Log> (Get Logs)
- [ ] GET /log/{log_id} (Get log)
- [ ] POST /user/location (address1, city, state, country, optional latitude, optional longitude, optional address2) -> location_id (Set user location)
- [ ] POST /org (name) -> org_id (Create Org)
- [ ] GET /project -> Array<Project> (Get projects the user has access to, optionally filtering by org they belong to)
- [ ] GET /filterItems GetProjectsAndOrgs() -> {"projects": Array<Project>, "orgs": Array<Org>}

#### Org auth (token and user that has accepted an org invite)

- [ ] POST /org/{id}/invite/{user_id} (from_user_id, to_user_id, optional project_id, optional org_id) -> invite_id (Create invitation to org)
- [ ] POST /org/{id}/invite/{invite_id} (invite_id) -> permitted_project_id (Accept/decline invitation to org)
- [ ] POST /org/{id}/location (address1, city, state, country, optional latitude, optional longitude, optional address2) -> location_id (Set Org Location)

### postgres database

Schema can be found in [the create tables sql file](sql/create_tables.sql).

### SDK implementations

SDKs should generally be designed so an instance of a struct or object can be created in the consumer's code, and authorization etc is handled for them.
The SDK should have an internal private function such as HandleRequest that either wraps each of the other endpoints. HandleRequest should do the following:

- Catch any 401 responses from the API, and if there is a 401 refresh the AuthorizationToken and retry. Raise an error if another 401 is returned.
- Raise for any other response codes > 400

### Admin Panel (Frontend)

#### Navigation

Should be configurable as a header or a sidebar. Should be minifiable if it's a sidebar.

- Logs
- Projects
- Organizations

#### Home/Log View

A logged in user should be taken straight to a feed of logs that is globally filterable in line with the server's GetLogs parameters.
The feed should poll the server for new results every 3 seconds if the user has "Poll" selected in their log-viewing screen.
There should be a sidebar of navigation that can be minified and should be displayed on the left.
Tab actions should select each filter dropdown or text input first.

#### Projects View

Should be a simple list of projects with some overall statistics for each.
Statistics should include the counts of each log status over a defineable amount of days.
Projects should be sortable by alphabet or by status count of any status.
There should be an extremely compact view where the user can fit a maximal amount of projects on the screen at a time.
User should be able to click a project to enter its project view.
Tab actions should flick through the distinct links to each project first.

#### Project View

Should feature the project name, a hideable description and a miniature log view filtered for that project in the bottom half of the page.
Should have a button for getting a new API Key for the project (the same endpoint would replace an existing one)

#### Organizations View

Should be a simple list of organizations with some overall statistics for each.
Statistics should include project count.
Organizations should be sortable by alphabet, join date (created_at of UserOrg) or by project count.
There should be an extremely compact view where the user can fit a maximal amount of organizations on the screen at a time.
User should be able to click a organization to enter its organization view.
Tab actions should flick through the distinct links to each organization first.

#### Organization View

Should organization name, hideable description and a list of projects that the user has access to that the organization collaborates on.
