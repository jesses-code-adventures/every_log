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

### postgres database

- User id, pii_id, created_at
- UserPii id, user_id, email, first_name, optional last_name, optional location_id (should have a relation to Location), optional mobile_number (string)
- AuthorizationToken id, created_at, unique api_key_id, token (will be a JWT)
- Org id, name, created_at, optional location_id (should have a relation to Location)
- Location id, address1, optional address2, city, state, country, optional latitude, optional longitude (combination of address1, address2, city, state, country should be unique)
- UserOrg id, user_id, org_id (combination of user_id and org_id should be unique), created_at (UserOrg is a many-to-many relation table linking User and Org), level (string literal of 'owner' | 'admin' | 'basic')
- Project id, user_id, name, created_at, optional description
- ProjectOrg id, org_id, created_at (ProjectOrg is a many-to-many relation table linking Project and Org)
- PermittedProject id, user_id, optional project_org_id (PermittedProject is a many-to-many relation table ensuring users who have permission to access projects are part of orgs that can access the project)
- Invite id, from_user_id, to_user_id, status (string literal, any of 'accepted' | 'pending' | 'declined' | 'expired'), optional project_id, optional org_id (there must be one of project_id or org_id)
- ApiKey id, unique permitted_project_id (relation to PermittedProject ensuring one key can be created per project permission), key (should be a string of any length)
- Process id, project_id (relation to Project), name, created_at (combination of name and project_id should be unique)
- Log id, created_at, project_id, level_id, optional process_id, message, optional traceback
- LogLevel id (this particular id should be an integer as it will be an enum with values 100-1000), value (string)

### server apis

All requests should require an Accept header of "application/json" so they can be expanded to return html later.
A 422 should be returned for:

- Missing Accept header
- Incorrectly typed data
- Missing required fields

A 401 should be returned for:

- Expired token (token created_at is > created_at + server defined expiration timedelta)
- Authorization token is not linked to an API key with permission to access the project.

The endpoints are as follows. Endpoint parameters assume user_id and an authorization_token are passed in the request headers.

- Authorize(user_id, api_key) -> authorization_token
- CreateUser(email, first_name, optional last_name, optional mobile_number) -> user_id
- SetUserLocation(address1, city, state, country, optional latitude, optional longitude, optional address2) -> location_id
- CreateOrg(name) -> org_id
- SetOrgLocation(address1, city, state, country, optional latitude, optional longitude, optional address2) -> location_id
- CreateProject(user_id, name, optional description) -> project_id
- CreateInvite(from_user_id, to_user_id, optional project_id, optional org_id) -> invite_id (api will need to validate that the user has permission to invite the other user. they will have to either be the creator of the project, or an admin/owner of the that collaborates on it.)
- AcceptInvite(invite_id) -> permitted_project_id
- CreateLog(level_id, project_id, message, optional process_id, optional traceback)
- GetLogs(optional projectId, optional level_id, optional process_id, optional org_id, optional from_datetime, optional to_datetime) -> Array<Log>
- GetProjects(optional org_id) -> Array<Project>

### SDK implementations

SDKs should generally be designed so an instance of a struct or object can be created in the consumer's code, and authorization etc is handled for them.
The SDK should have an internal private function such as HandleRequest that either wraps each of the other endpoints. HandleRequest should do the following:

- Catch any 401 responses from the API, and if there is a 401 refresh the AuthorizationToken and retry. Raise an error if another 401 is returned.
- Raise for any other response codes > 400

### Admin Panel

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
