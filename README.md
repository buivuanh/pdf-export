# pdf-export

## Structure

`/html`: html, template, css, json data files.

`output`: contain output pdf files.

`slide`: slide.

`main.go`: logic code.

## Quick start

Build binary:

`go build`

### Render static html:

```
./pdf-export e -p=html -t=static_data.html
```

`-p`: workdirectory

`-t`: main html file with a path is workdirectory/main.html, eg: html/static_data.html



### Render html with given json file:

```
./pdf-export e -p=html -t=template.html -d=data.json
```

`-p`: work directory

`-t`: main html file with a path is workdirectory/main.html, eg: html/template.html

`-d`:  json file with a path is workdirectory/data.json, eg: html/data.json.html

To add images or any file into html/css, you can:
- Inline file's data, e.g:  ```background: url('data:image/svg+xml;utf8,%3Csvg%20width%3D%22260%2``` 
- Or put into `/html` directory, a static file server will run at port 8181 when u run `./pdf-export`. E.g: `http://localhost:8181/logo.svg`

## Run Slide local

```shell
go get -u golang.org/x/tools/present

present
```

Open http://127.0.0.1:3999 in chrome. 