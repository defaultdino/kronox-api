# Middleware

### Purpose

It is **not** the responsibility of kronox-api to validate or refresh the users `session_id` (JSESSIONID) before calling any endpoint. It simply forwards and requires a `session_id` to exist before calling any authorized endpoint.
