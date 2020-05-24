# Remote File System [golang, svelte.js]

## Requirements
- GoLang
- NodsJS 10+

## Start the app
### API
Inside /api, ```go run main.go```
Lanches a golang server on http://localhost:8080

NB: 
- main.go is a (long) single file for development reasons, could not develop properly on VSCOde with multiple golang files.
- Don't touch the initial files (file1.txt + file.txt) inside /api/temp-files.

### Front-end
Inside /front, run ```npm run dev``` 
Launches a node server on http://localhost:5000. You can access the wep app at this location.