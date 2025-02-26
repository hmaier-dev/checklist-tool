package helper

import (
	"encoding/json"
	"log"
  "fmt"
  "github.com/hmaier-dev/checklist-tool/internal/checklist"
)


// I use this function to add the IMEI to every ChecklistItem
// so I can reference it in the template.
// I know, this is not efficient but it works.
func AddDataToEveryEntry(toAdd string, jsonArray []*checklist.ChecklistItem ){
  for _, item := range jsonArray{
    item.IMEI = toAdd
    if len(item.Children) > 0 {
			AddDataToEveryEntry(toAdd, item.Children)
		}
  }

}

func prettyPrintJSON(jsonArray any){
  //Print the unmarshaled JSON
  jsonData, err := json.MarshalIndent(jsonArray, "", "  ")
  if err != nil {
          log.Println("JSON marshal error:", err)
          return
  }
  fmt.Println(string(jsonData))

}


func ChangeCheckedStatus(newItem checklist.ChecklistItem, oldChecklist []*checklist.ChecklistItem){
  for _, item := range oldChecklist{
    if newItem.Task == item.Task{
      item.Checked = newItem.Checked     
      return
    }
    if len(item.Children) > 0 {
			ChangeCheckedStatus(newItem, item.Children)
		}
  }
}
