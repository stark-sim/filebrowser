// 删除local中list对象中的name项
export function deleteListItem(name) {
  let tempList = JSON.parse(localStorage.getItem("list"));
  delete tempList[name];
  localStorage.setItem("list", JSON.stringify(tempList));
}
