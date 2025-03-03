package helper

import (
  "github.com/hmaier-dev/checklist-tool/internal/structs"
)

// I use this function to add the Path to every ChecklistItem
// so I can reference it in the template.
// I know, this is not efficient but it works.
func AddDataToEveryEntry(toAdd string, clArray []*structs.ChecklistItem ){
  for _, item := range clArray{
    item.Path = toAdd
    if len(item.Children) > 0 {
			AddDataToEveryEntry(toAdd, item.Children)
		}
  }

}

func ChangeCheckedStatus(newItem structs.ChecklistItem, oldChecklist []*structs.ChecklistItem){
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
