// 删除local中list对象中的name项
export function deleteListItem(name) {
  let tempList = JSON.parse(localStorage.getItem("list"));
  delete tempList[name];
  localStorage.setItem("list", JSON.stringify(tempList));
}

//代表302占用通道的状态
export function changeListItemstatus(name) {
  let tempList = JSON.parse(localStorage.getItem("list"));
  tempList[name].process += 1;
  localStorage.setItem("list", JSON.stringify(tempList));
}
