Possible approaches for handling dynamic paths for HTTP servers.

Where paths like `/foo/bar/%s/baz/%s/qux` (the only path implemented here) can be used with the stdlib.

Uses two implementation :`ServeMux` and a custom router that associates routes with handlers for those routes.

i.e

```bash
curl http://localhost:8080/foo/bar/param123/baz/param45632131/qux
Path parameters received:
First parameter: param123
Second parameter: param45632131
```

Defaults to `ServeMux` implementation; pass `-customRouter` for that approach.
