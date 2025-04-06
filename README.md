# gogetter

> An incremental HTTP client 

1. Start small and easily write requests;
2. Save and variabilize your most used requests if needed;
3. Scale and share your requests with your team in a git-friendly format.

## Language specs

### Basics

A request MUST start with the HTTP method followed by the URL. 

Valid request:

```
GET https://pkg.go.dev
```

Invalid requests:

```
https://pkg.go.dev
```

```
GET
```

A request COULD contain any number of spaces and carriage returns between elements.

Valid requests:

```
GET       https://pkg.go.dev
```

```
GET

       https://pkg.go.dev
```

### URL search params

A request COULD contain search parameters directly in the URL.

```
GET https://github.com/ThomasFerro/gogetter/issues?q=feat
```

A request COULD also contain search parameters that will be encoded using the following syntax:

```
GET https://github.com/ThomasFerro/gogetter/issues
q=?feat
```

```
GET https://github.com/ThomasFerro/gogetter/issues
q=?"my search"
```

Parameters provider in the URL directly WILL be overwrote by the same parameter provided later on in the request.

### Headers

A request COULD contain headers:

```
GET https://api.com/list
x-api-key=:my-api-key
```

### Body

A request COULD end with a body. The type of body will be interpreted from the definition and will be set in the `Content-Type` header).

The two types of valid bodies are:

1. JSON (producing a `application/json` content type)
```
POST https://api.com/
{ "item": { "name": "First item", "id": 10 } }
```  
2. Key/value pairs (producing a `multipart/form-data` content type)
```
POST https://api.com/
name="First item" id=10
```  

A request MUST contain up to one body, any request with more than one body definition will be considered invalid.

### Variables

A request COULD contain variables following [Go's templating format](https://pkg.go.dev/text/template).

