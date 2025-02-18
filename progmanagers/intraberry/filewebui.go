/*
Browse files with browser.

# Template based

Really bad idea to have this feature on production. Great when testing and exploring
*/
package main

import (
	"fmt"
	"html/template"
)

const basicDirHTMLTemplateRAW = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Listing {{.De.Path}} {{.UpdatedTimeAndDate}}</title>
    <style>

    body {
            margin: 0;
            font-family: Arial, sans-serif;
            display: flex;
        }

         .side-menu {
            width: 250px;
            background-color: #333;
            padding: 20px;
            box-sizing: border-box;
            color: white;
        }

        .side-menu ul {
            list-style-type: none;
            padding: 0;
            margin: 0;
        }

        .side-menu ul li {
            margin: 10px 0;
        }

        .side-menu ul li a {
            color: white;
            text-decoration: none;
            font-size: 16px;
        }

        .side-menu ul li a:hover {
            color: #ddd;
        }

        .side-menu ul li ul {
            margin-left: 20px;
            display: block;
        }

            
        .content {
            flex-grow: 1;
            padding: 20px;
            background-color: #f4f4f4;
        }

        .column {
            flex: 1;
            padding: 20px;
            box-sizing: border-box;
        }

        .column:first-child {
            background-color: #e0e0e0;
        }

        .column:last-child {
            background-color: #d0d0d0;
        }

    .dirtag {
        display: inline-flex;
        align-items: center;
        background-color: #007BFF;
        color: white;
        font-size: 12px;
        font-weight: bold;
        padding: 4px 10px;
        border-radius: 20px;
        border: none;
        gap: 5px;
    }
    .devtag {
        display: inline-flex;
        align-items: center;
        background-color:rgb(255, 255, 0);
        color: rgb(0, 0, 0);
        font-size: 12px;
        font-weight: bold;
        padding: 4px 10px;
        border-radius: 20px;
        border: none;
        gap: 5px;
    }
    </style>

</head>
<body>

<div class="side-menu">
        {{.SideTree}}
    </div>

    <div class="content">

    <div class="column">
    <h1>Directory Listing: {{.De.Path}} (total:{{.De.Size}})</h1>
    {{.UpdatedTimeAndDate}}
    <br>
    {{if .De.Dirs}}
    <h2>Dirs</h2>

    {{range .De.Dirs}}
        <span class="dirtag"><a href="{{.PathLink}}"> {{.Name}}</a></span>
    {{end}}
    {{end}}

    {{if .De.Files}}
    <h2>Files</h2>
    <table>
        <tr><th>FileName</td><th>Size</th><th>Mime</th><th>Time</th></tr>
        {{range .De.Files}}
            <tr>
                <td><a href="{{.PathLink}}">{{.Name}}</a></td>
                <td><a href="{{.PreviewLink}}"> {{.Size}}</a></td> 
                <td>{{.Resolved.MimeType}}</td> 
                <td>{{.ModTimeFormatted}}</td></tr>
        {{end}}
    </table>
    {{end}}

    {{if .De.Links}}
    <h2>Links</h2>
    <ul>
        {{range .De.Links}}
        <li><strong>Link:</strong> {{.Path}} -> {{.LinkTo.Path}}</li>
        {{end}}
    </ul>
    {{end}}

    {{if .De.DeviceFiles}}
    <h2>DeviceFiles</h2>
    {{range .De.DeviceFiles}}
        <span class="devtag"> {{.Name}}</span>
    {{end}}


    {{end}}








    {{if .De.NamedPipes}}
        <h2>NamedPipes</h2>
        <ul>
            {{range .De.NamedPipes}}
            <li><strong>Named Pipe:</strong> {{.Path}}</li>
            {{end}}
        </ul>
    {{end}}
    {{if .De.Sockets}}
        <h2>Sockets</h2>
        <ul>
            {{range .De.Sockets}}
                <li><strong>Socket:</strong> {{.Path}}</li>
            {{end}}
        </ul>
    {{end}}
    </div>

 
    {{if .PreviewText}}
    <div class="column">
        <pre>{{.PreviewText}}</pre>
    </div>
    {{end}}

    </div>
</body>
</html>
`

var basicDirHTMLTemplate *template.Template

func initDirGenerator() error {
	// Parse the template
	var err error
	basicDirHTMLTemplate, err = template.New("directory").Parse(basicDirHTMLTemplateRAW)
	if err != nil {
		return fmt.Errorf("internal template error %w", err)
	}
	return nil
}
