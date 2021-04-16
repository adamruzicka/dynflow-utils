# escape-csv

The "CSV" exports collected by sosreport are not proper CSVs. They are a comma
separated list of values, but the values themselves can be JSON objects which
are escaped in any way. This tool tries to detect those fields and escape them
to create a "proper" CSV.

## Usage

```shell
# cat dynflow_actions.broken
4ceffc10-cb6e-432f-a9c5-578935e69782,197,,,196,Actions::Pulp::Orchestration::Repository::RefreshIfNeeded,{"remote_user":"admin","remote_cp_user":"admin","current_request_id":null,"current_timezone":"Pacific/Auckland","current_user_id":1,"current_organization_id":1,"current_location_id":null},,263,,

# escape-csv < dynflow_actions.broken > dynflow_actions.csv

# cat dynflow_actions.csv
4ceffc10-cb6e-432f-a9c5-578935e69782,197,,,196,Actions::Pulp::Orchestration::Repository::RefreshIfNeeded,"{""remote_user"":""admin"",""remote_cp_user"":""admin"",""current_request_id"":null,""current_timezone"":""Pacific/Auckland"",""current_user_id"":1,""current_organization_id"":1,""current_location_id"":null}",,263,,
```
