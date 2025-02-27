param(
  [switch]$buildDocker = $False,
  [switch]$buildGo = $False,
  [switch]$buildCss = $False,
  [switch]$all = $False

)

if($all){
  $buildGo = $true
  $buildCss = $true
}

if($buildGo){
  Remove-Item .\bin\checklist-tool.exe
	go build -o .\bin\checklist-tool.exe main.go
  Get-ChildItem .\bin

}
if($buildDocker){
	docker build -t "checklist-tool" .
}
if($buildCss){
	tailwindcss -i ./static/base.css -o ./static/style.css
}
