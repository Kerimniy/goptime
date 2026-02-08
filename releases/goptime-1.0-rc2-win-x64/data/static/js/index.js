  if (logined=="true"){
    let n = document.getElementsByName("auth")
    for (let e of n){
      e.style.display=""
    }
    let lg = document.getElementById("login")
    lg.lastElementChild.innerText="Logout"
    lg.firstElementChild.src="/static/svg/logout.svg"
    lg.addEventListener("click",()=>{
      fetch("/logout/")
      window.location.reload()
    })
  }
  else{
    let lg = document.getElementById("login")
   
    lg.addEventListener("click",()=>{
      window.location.href="/login/"
    })
  }
  
  document.title = server_data["title"]
  document.getElementById("title").innerText = server_data["title"]
  if (server_data["md"]===undefined||server_data["md"].trim()==""){
    document.getElementById("md-content").parentElement.remove()
  }
  else{
    document.getElementById("md-content").innerHTML =marked.parse(server_data["md"])
  }

  
  let md_height = 0
  try{
    md_height= document.getElementById("md-content").parentElement?.offsetHeight+parseFloat(getComputedStyle(document.getElementById("md-content").parentElement).marginTop)*2 || 0
  }
  catch (error){

  }
  document.getElementById("chart").style.minHeight=`calc(67.5vh - ${md_height}px)`
 

  let monitors = {}
  let m_length = 0
  for (let el of monitors_raw){
    m_length++
    if (!monitors.hasOwnProperty(el["group"])){
      monitors[el["group"]] = [el]
    }
    else{
      monitors[el["group"]].push(el)
    }
  }

  if (monitors==undefined || m_length==0){
    window.location.href="/admin/"
  }


  let chart = document.getElementById("chart")
  let current_group

  

  for (let group in monitors){

    current_group = document.createElement("div")
    current_group.classList.add("group-container")
    let label = document.createElement("div")
    label.classList.add("group-label")
    label.innerText = group

    current_group.appendChild(label)

    for (let service of monitors[group]){

      let service_container = document.createElement("div")
      service_container.classList.add("service-container")
      service_container.id = "service-"+service["name"]

      let service_grid = document.createElement("div")
      service_grid.classList.add("service-grid")

      let service_uptime = document.createElement("div")
      service_uptime.classList.add("service-uptime")
      service_uptime.style.background = "rgb(175,52,0)"
      
      service_uptime.style.background = p2rgb(Number(service["uptime"]))
      let service_uptime_text = document.createElement("p")
      service_uptime_text.innerText = Math.round(Number(service["uptime"]*100))+"%"

      let service_name = document.createElement("div")
      service_name.classList.add("service-name")

      let service_name_text = document.createElement("p")
      service_name_text.innerText = service["name"]

      let uptimebar = document.createElement("div")
      uptimebar.classList.add("uptimebar")

      let shimmer_container = document.createElement("div")
      shimmer_container.classList.add("shimmer-container")
      shimmer_container.style.display = "none"
      let shimmer_overlay = document.createElement("div")
      shimmer_overlay.classList.add("shimmer-overlay")


      shimmer_container.appendChild(shimmer_overlay)

      service_uptime.appendChild(service_uptime_text)
      service_name.appendChild(service_name_text)

      service_grid.appendChild(service_uptime)
      service_grid.appendChild(service_name)
      service_grid.appendChild(uptimebar)

      service_container.appendChild(service_grid)
      service_container.appendChild(shimmer_container)

      current_group.appendChild(service_container)

      for (let i=0; i<30-service["checks"].length;i++){

        
        let check_el = document.createElement("div")

        check_el.style.background = "var(--grey)"


        let tool_tip 

        check_el.addEventListener("mouseenter",(e)=>{
          tool_tip = document.createElement("div")
          tool_tip.classList.add("js-tooltiptext")
          tool_tip.style.top = e.clientY+15+"px"
          tool_tip.style.left = e.clientX+15+"px"

          tool_tip.innerText = "No data"

          document.body.appendChild(tool_tip)
        })
        check_el.addEventListener("mousemove",(e)=>{

          tool_tip.style.top = e.clientY+15+"px"
          tool_tip.style.left = e.clientX+15+"px"
        })
        check_el.addEventListener("mouseleave",()=>{
          tool_tip.remove()
        })


        uptimebar.appendChild(check_el)
      }

      for (let i=service["checks"].length-1; i>-1;i--){
        let check_el = document.createElement("div")

        if (service["checks"][i]["ok"]==0){
          check_el.style.background = "var(--red)"
        }
        else{
          check_el.style.background = "var(--green)"
        }

        let tool_tip 

        check_el.addEventListener("mouseenter",(e)=>{
          tool_tip = document.createElement("div")
          tool_tip.classList.add("js-tooltiptext")
          tool_tip.style.top = e.clientY+15+"px"
          tool_tip.style.left = e.clientX+15+"px"

          tool_tip.innerText = timeConverter(Number(service["checks"][i]["timestamp"]))

          document.body.appendChild(tool_tip)
        })
        check_el.addEventListener("mousemove",(e)=>{

          tool_tip.style.top = e.clientY+15+"px"
          tool_tip.style.left = e.clientX+15+"px"
        })
        check_el.addEventListener("mouseleave",()=>{
          tool_tip.remove()
        })

        uptimebar.appendChild(check_el)
      }
    }
  
    
    chart.appendChild(current_group)
  
  }  

  const rate = 300
  let time = rate


  const counter_div = document.getElementById("counter")

  counter_div.innerText = `Refresh in ${Math.floor(time/60)}m ${time%60}s`

  setInterval(()=>{time--; if (time<0){time = rate; update()}; counter_div.innerText = `Refresh ${Math.floor(time/60)}m ${time%60}s`; },1000)




  function update(){

    for (el of document.querySelectorAll(".service-container")){
      el.lastElementChild.style.display = "block"
      el.firstElementChild.style.display = "none"
    }

    fetch(`/get_info_from?time=${Math.round(Date.now()/1000)-rate}`, {method: "GET"})
    .then((rsp)=>{


      if (rsp.ok){
        return rsp.json()
      }
      else{
        console.error(rsp.status)
        for (el of document.querySelectorAll(".service-container")){
           setTimeout(()=>{
            el.lastElementChild.style.display = "none"
            el.firstElementChild.style.display = ""},
            1000
          )
        }
      }
    }) 
    .then((raw_data)=>{

      let tips = document.querySelectorAll(".js-tooltiptext")

      for (let tip of tips){
        tip.remove()
      }

      
      for (let el of raw_data){

      let service_container = document.getElementById("service-"+el["name"])
      if (service_container===undefined){
        continue
      }

      service_container.lastElementChild.style.display = "block"
      service_container.firstElementChild.style.display = "none"

      let service_grid = service_container.firstElementChild

      service_grid.firstElementChild.style.background = p2rgb(Number(el["uptime"]))
      service_grid.firstElementChild.firstChild.innerText = Math.round(Number(el["uptime"])*100)+"%"


      let uptimebar = service_grid.lastElementChild

      for (let i=0; i<el["checks"].length;i++){
          uptimebar.firstElementChild.remove()
      }


      for (let i=el["checks"].length-1; i>-1;i--){
        let check_el = document.createElement("div")

        if (el["checks"][i]["ok"]==0){
          check_el.style.background = "var(--red)"
        }
        else{
          check_el.style.background = "var(--green)"
        }

        let tool_tip 

        check_el.addEventListener("mouseenter",(e)=>{
          tool_tip = document.createElement("div")
          tool_tip.classList.add("js-tooltiptext")
          tool_tip.style.top = e.clientY+15+"px"
          tool_tip.style.left = e.clientX+15+"px"

          tool_tip.innerText = timeConverter(Number(el["checks"][i]["timestamp"]))

          document.body.appendChild(tool_tip)
        })
        check_el.addEventListener("mousemove",(e)=>{

          tool_tip.style.top = e.clientY+15+"px"
          tool_tip.style.left = e.clientX+15+"px"
        })
        check_el.addEventListener("mouseleave",()=>{
          tool_tip.remove()
        })

        uptimebar.appendChild(check_el)
      
      }



      setTimeout(()=>{
        service_container.lastElementChild.style.display = "none"
        service_container.firstElementChild.style.display = ""},
        1000
      )

    }

    })

    
        
  }

  function p2rgb(v){
    let g
    let r

    v-=0.25
    if (v<0){
      v=0
    }

    if (v<0.5){

    r = 255
    g=(v*2)*255
    }
    else{
      r=(1-(v-0.5)*2)*255
      if (r<0){
        r=0
      }
      g = 255
    }
    return `rgb(${r*0.9},${g*0.9},0)`
  }

  function timeConverter(UNIX_timestamp){
    var a = new Date(UNIX_timestamp * 1000);
    var months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
    var year = a.getFullYear();
    var month = months[a.getMonth()];
    var date = a.getDate();
    var hour = a.getHours();
    var min = a.getMinutes();
    var sec = a.getSeconds();
    var time = date + ' ' + month + ' ' + year + ' ' + hour.toString().padStart(2, '0') + ':' + min.toString().padStart(2, '0') + ':' + sec.toString().padStart(2, '0') ;
    return time;
  }

