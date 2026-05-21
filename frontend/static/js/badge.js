let badges = document.querySelectorAll("[data-badge-url]")
let timestamp = Date.now();
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


for (let badge of badges){
    let url = badge.dataset.badgeUrl
    let start = performance.now()

     fetch(`${url}?time=${String(timestamp)}`, {
      method: "get",
    }).then(r=>{
      return r.json()
    }).then(data1=>{

      let ping = (performance.now()-start)/1000


      let data = JSON.parse(atob(data1.content))


      let unixTime = Math.floor(Date.now() / 1000);
      let elapsed = unixTime - data.reset

      let uptime = Math.floor(data.uptime/data.interval)/Math.floor(elapsed/data.interval-ping)
      if (uptime==Infinity){
        uptime=0
      }
      else if (uptime>1){
        uptime=1  
      }

      let badgSVG = `
      
        <svg xmlns="http://www.w3.org/2000/svg" width="130" height="20" role="img" aria-label="h24uptime">
            <title>Uptime (24h)</title>
            <linearGradient id="s" x2="0" y2="100%">
                <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
                <stop offset="1" stop-opacity=".1"/>
            </linearGradient>
            <clipPath id="r">
                <rect width="125" height="20" rx="3" fill="#fff"/>
            </clipPath>
            <g clip-path="url(#r)">
                <rect width="80" height="20" fill="#4d4d4dff"/>
                <rect x="77" width="50" height="20" fill=${p2rgb(uptime)}>
                <rect width="130" height="20" fill="url(/)"/>
            </g>
            <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">
                <text aria-hidden="true" x="395" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="700">Uptime (24h)</text>
                <text x="395" y="140" transform="scale(.1)" fill="#fff" textLength="700">Uptime (24h)</text>
                <text aria-hidden="true" x="1000" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)">${uptime*100}%</text>
                <text x="1000" y="140" transform="scale(.1)" fill="#fff" >${uptime*100}%</text>
            </g>
        </svg>
      
      `

      badge.innerHTML=badgSVG

    
    })

}