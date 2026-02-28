

let statusbar_container = document.createElement("div")
statusbar_container.classList.add("status-bar-container")
statusbar_container.id="st_container"
document.body.appendChild(statusbar_container)

  function createStatus(st){
    let statusBar = document.createElement("div")
    let text = document.createElement("p")
    let close = document.createElement("a")
    statusBar.classList.add("status-bar")
    text.innerText = st || "OK"
    close.innerText = "🞫"
    statusBar.appendChild(text)
    statusBar.appendChild(close)

    if (st==""||st==undefined){
      statusBar.style.background="rgb(50,150,0)"
    }

    close.addEventListener("click",()=>{
      statusBar.style.transform = 'translateX(100%)'

      setTimeout(()=>{statusBar.remove()},200)
    })

    setTimeout(()=>{statusBar.style.transform = 'translateX(0%)'},200)
    setTimeout(()=>{statusBar.style.transform = 'translateX(100%)';setTimeout(()=>{statusBar.remove()},200)},5000)
    statusbar_container.appendChild(statusBar)

  }