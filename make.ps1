param(
  [switch]$buildDocker = $False,
  [switch]$buildGo = $False
)

if($buildGo){
	mkdir -p ./bin
	go build -o bin/checklist-tool.exe main.go

}
if($buildDocker){
	docker build -t "checklist-tool" .
}
