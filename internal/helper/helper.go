package helper

import (
  "github.com/hmaier-dev/checklist-tool/internal/checklist"
)

// I use this function to add the IMEI to every ChecklistItem
// so I can reference it in the template.
// I know, this is not efficient but it works.
func AddDataToEveryEntry(toAdd string, clArray []*checklist.ChecklistItem ){
  for _, item := range clArray{
    item.IMEI = toAdd
    if len(item.Children) > 0 {
			AddDataToEveryEntry(toAdd, item.Children)
		}
  }

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
