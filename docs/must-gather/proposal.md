# ISV-CLI Must Gather

ISV-CLI is for ISV managed service operator.

## High-Level Understanding 
 **about supporting for ISV managed service operator** 
- When managed service operators have issues, customers(end-user/Red Hat support engineer) need to engage with ISV directly.
- ISV needs to gather data for debugging.

## Assumption
- Customers have only Admin role.
- Each managed service operator must be in a specific namespace.
- ISV can gather all data in the specific namespace except Secret.

## OC CLI Must-Gather
- OC cli already have `oc adm must-gather` and `oc adm inspect commands` to gather necessary data.
- However, it requires `cluster-admin` role to execute the command.

## ISV-CLI Must-Gather
- ISV-CLI customize OC CLI Must-Gather to meet the requirements.
- ISV-CLI provides the following features:
  - login
  - logout
  - must-gather
- Anybody who has `admin user` role can execute the must-gather command.
- Output format should be 
  - plain text files which is the same as `oc cli must-gather`
  - tarball format.
- event-filter.html will be provided.

