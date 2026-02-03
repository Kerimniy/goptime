let statusBar = document.getElementById("status-bar")

let title_input = document.getElementById("title_input")
let md = document.getElementById("md")
let finput = document.getElementById("finput")
let finputbutton = document.getElementById("finputbutton")

title_input.value = server_data["title"]
md.value = server_data["md"]

let monitors_settings = document.getElementById("monitors_settings")


for (let e_data of data){
  createMonitorBlock(e_data)
}


function createMonitorBlock(e_data){
  let url_container = document.createElement("div")
  let url_label = document.createElement("div")
  let url_textarea_div = document.createElement("div")
  let url_textarea = document.createElement("textarea")

  let name_container = document.createElement("div")
  let name_label = document.createElement("div")
  let name_textarea_div = document.createElement("div")
  let name_textarea = document.createElement("textarea")

  let group_container = document.createElement("div")
  let group_label = document.createElement("div")
  let group_textarea_div = document.createElement("div")
  let group_textarea = document.createElement("textarea")

  let interval_container = document.createElement("div")
  let interval_label = document.createElement("div")
  let interval_input_div = document.createElement("div")
  let interval_input = document.createElement("input")

  let timeout_container = document.createElement("div")
  let timeout_label = document.createElement("div")
  let timeout_input_div = document.createElement("div")
  let timeout_input = document.createElement("input")

  let grid = document.createElement("div")
  grid.classList.add("createmonitorgrid")

  let button_container = document.createElement("div")
  let button_text_dummy = document.createElement("div")
  let button = document.createElement("button")
  
  let delete_button_container = document.createElement("div")
  let delete_button_text_dummy = document.createElement("div")
  let delete_button = document.createElement("button")

  url_label.style.textAlign = "center"
  name_label.style.textAlign = "center"
  group_label.style.textAlign = "center"
  interval_label.style.textAlign = "center"
  timeout_label.style.textAlign = "center"


  url_textarea.classList.add("textinput")
  name_textarea.classList.add("textinput")
  group_textarea.classList.add("textinput")
  interval_input.classList.add("textinput")
  timeout_input.classList.add("textinput")

  button.classList.add("buttonSettings")
  button.style.gridColumn = "1 / span 2"
  button.innerText = "Update"
  
  delete_button.classList.add("buttonSettings")
  delete_button.style.gridColumn = "1 / span 2"
  delete_button.innerText = "Delete"
  delete_button.style.background = "var(--danger)"

  button_text_dummy.innerText=" "
  button_text_dummy.style.whiteSpace = "pre"

  interval_input.type = "number"
  timeout_input.type = "number"

  url_textarea_div.appendChild(url_textarea)
  url_container.appendChild(url_label)
  url_container.appendChild(url_textarea_div)

  name_textarea_div.appendChild(name_textarea)
  name_container.appendChild(name_label)
  name_container.appendChild(name_textarea_div)

  group_textarea_div.appendChild(group_textarea)
  group_container.appendChild(group_label)
  group_container.appendChild(group_textarea_div)

  interval_input_div.appendChild(interval_input)
  interval_container.appendChild(interval_label)
  interval_container.appendChild(interval_input_div)

  timeout_input_div.appendChild(timeout_input)
  timeout_container.appendChild(timeout_label)
  timeout_container.appendChild(timeout_input_div)

  button_container.appendChild(button_text_dummy)
  button_container.appendChild(button)

  delete_button_container.appendChild(delete_button_text_dummy)
  delete_button_container.appendChild(delete_button)

  url_label.innerText = "Url"
  name_label.innerText = "Name"
  group_label.innerText = "Group"
  interval_label.innerText = "Interval"
  timeout_label.innerText = "Timeout"

  url_textarea.value = e_data["url"]
  name_textarea.value = e_data["name"]
  group_textarea.value = e_data["group"]
  interval_input.value = Number(e_data["interval"])
  timeout_input.value = Number(e_data["timeout"])

  grid.appendChild(url_container)
  grid.appendChild(name_container)
  grid.appendChild(group_container)
  grid.appendChild(interval_container)
  grid.appendChild(timeout_container)
  grid.appendChild(button_container)
  grid.appendChild(document.createElement("div"))
  grid.appendChild(document.createElement("div"))
  grid.appendChild(document.createElement("div"))
  grid.appendChild(delete_button_container)

  grid.dataset.cname = name_textarea.value

  monitors_settings.appendChild(grid)

  button.addEventListener("click",()=>{
    let res = {"cname":grid.dataset.cname, "url": url_textarea.value, "name": name_textarea.value, "group": group_textarea.value, "interval":interval_input.value,"timeout": timeout_input.value } 

     let st;
    fetch(`/update-monitor/`, {method: 'POST',headers: {'Content-Type': 'application/json'},  body: JSON.stringify(res)})
      .then(resp => {
           st = resp.ok
        if (resp.ok){
          statusBar.style.backgroundColor = "rgb(25,168,0)"
        }
        else{
          statusBar.style.backgroundColor = "rgb(168,0,0)"
        }
        return resp.text()
      })
       .then((resp)=>{
      if (st){
        statusBar.firstElementChild.innerText = "OK"
      }  
      else{
        statusBar.firstElementChild.innerText = `Error: ${resp}`
      }

      statusBar.style.transform = "translateX(0%)"
      setTimeout(()=>{statusBar.style.transform = "translateX(100%)"},3000)
    })

  })
 
  delete_button.addEventListener("click",()=>{
    let cname = grid.dataset.cname 
    
    let st;

    fetch(`/delete-monitor/`, {method: 'POST',headers: {'Content-Type': 'text/plain'},  body: cname})
      .then(resp => {
          st = resp.ok
        if (resp.ok){
          statusBar.style.backgroundColor = "rgb(25,168,0)"
        }
        else{
          statusBar.style.backgroundColor = "rgb(168,0,0)"
        }
        return resp.text()
      })
      .then((resp)=>{
      if (st){
        statusBar.firstElementChild.innerText = "OK"
        grid.remove()
      }  
      else{
        statusBar.firstElementChild.innerText = `Error: ${resp}`
      }

      statusBar.style.transform = "translateX(0%)"
      setTimeout(()=>{statusBar.style.transform = "translateX(100%)"},3000)
    })

  })  
}


function createMonitorRequest(){
    let create_url = document.getElementById("create-url")
    let create_name = document.getElementById("create-name")
    let create_group = document.getElementById("create-group")
    let create_interval = document.getElementById("create-interval")
    let create_timeout = document.getElementById("create-timeout")

    let res = {"url":create_url.value, "name":create_name.value, "group": create_group.value, "interval": Number(create_interval.value), "timeout": Number(create_timeout.value)}
     let st;

    fetch(`/create-monitor/`, {method: 'POST',headers: {'Content-Type': 'application/json'},  body: JSON.stringify(res)})
      .then(resp => {
        st = resp.ok
        if (resp.ok){
          createMonitorBlock(res)
          statusBar.style.backgroundColor = "rgb(25,168,0)"
        }
        else{
          statusBar.style.backgroundColor = "rgb(168,0,0)"
        }
        return resp.text()
      })
      .then((resp)=>{
      if (st){
        statusBar.firstElementChild.innerText = "OK"
      }  
      else{
        statusBar.firstElementChild.innerText = `Error: ${resp}`
      }

      statusBar.style.transform = "translateX(0%)"
      setTimeout(()=>{statusBar.style.transform = "translateX(100%)"},3000)
    })
  }

  function imitatefinput(){
    finput.click()
  }

  function setfilename(){
    finputbutton.innerText = finput.files[0].name || "Upload"
  }

  function chengeServerSettings(){
    let data = new FormData()
    data.append("title", title_input.value)
    data.append("md", md.value)
    if (finput.files.length>0){
      data.append("image", finput.files[0])
    }

    fetch(`/update-server/`, {method: 'POST',  body: data})
      .then(resp => {
        st = resp.ok
        if (resp.ok){
          statusBar.style.backgroundColor = "rgb(25,168,0)"
        }
        else{
          statusBar.style.backgroundColor = "rgb(168,0,0)"
        }
        return resp.text()
      })
      .then((resp)=>{
      if (st){
        statusBar.firstElementChild.innerText = "OK"
      }  
      else{
        statusBar.firstElementChild.innerText = `Error: ${resp}`
      }

      statusBar.style.transform = "translateX(0%)"
      setTimeout(()=>{statusBar.style.transform = "translateX(100%)"},3000)
    })
  
  }