param(
  [switch]$buildDocker = $False,
  [switch]$buildGo = $False
)

if($buildGo){
  Remove-Item .\bin\checklist-tool.exe
	go build -o .\bin\checklist-tool.exe main.go

}
if($buildDocker){
	docker build -t "checklist-tool" .
}
