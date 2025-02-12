// Initialize Ractive.js
const ractive = new Ractive({
  el: '#app',
  template: '#template',
  data: {
    page:"stdout",
    stdout: 'Waiting for output...',
    cpuLoad: 'Fetching CPU load...',
    diskStatus: 'Fetching disk status...'
  },
  partials:{
    pageStdout:"<h3>Stdout</h3><pre>{{stdout}}</pre>",
    pageStderr:"<h3>Stderr</h3><pre>{{stderr}}</pre>",
    pageChrashmsg:"<h3>Crash Messages</h3><pre>{{chrashmsg}}</pre>",
    pageCpu:"<h3>CPU Load</h3><p>{{cpuLoad}}</p>",
    pageDisk:"<h3>Disk Status</h3><p>{{diskStatus}}</p>",
    pageEnv:"<h3>Environmental values</h3><textarea rows='30' cols='80' value={{env}}></textarea><button on-click=\"['setenv']\">SET envVariables</button> {{envFeedback}}",
    pageKmsg:"<h3>Kernel Message</h3><pre>{{kmsg}}</pre>",
    pageMounts:`<h3>Mounts</h3>
    <table border>
        <tr>
        <th>Device</th><th>MountPoint</th><th>Filesystem</th><th>Options</th><th>Free</th><th>Capacity</th><th></th>
        </tr>
      {{#each mounts}}
        <tr>
        <td>{{Device}}</td><td>{{MountPoint}}</td><td>{{Filesystem}}</td><td>{{Options}}</td><td>{{Free}}</td><td>{{Capacity}}</td><td>
        <progress value="{{Capacity-Free}}" max="{{Capacity}}">  </progress></td>
        </tr>
      {{/each}}
    </table>`,
    pageBlockdevices:`<h3>Block devices</h3><pre>{{blockdevices}}</pre><br>File handle usage:{{filehandleusage}}`,
    pageProcstat:`<h3>Proc stats</h3>
    <ul>
      <li>Interrupts:{{procstat.Interrupts}}</li>
	    <li>Ctxt:{{procstat.Ctxt}}</li>
	    <li>BootEpoch:{{procstat.BootEpoch}}</li>
	    <li>Processes:{{procstat.Processes}}</li>
	    <li>ProcsRunning:{{procstat.ProcsRunning}}</li>
	    <li>ProcsBlocked:{{procstat.ProcsBlocked}}</li>
	    <li>SoftIRQTotal:{{procstat.SoftIRQTotal}}</li>
	    <li>SoftIRQs:{{JSON.stringify(procstat.SoftIRQs)}}</li>
	    <li>InterruptsByInt:{{JSON.stringify(procstat.InterruptsByInt)}}</li>
    </ul>

    <h4>CPU</h4>
    <table border>
    <tr>
      <th>CPU</th>
      <th>User</th>
      <th>Nice</th>
      <th>System</th>
      <th>Idle</th>
      <th>IOWait</th>
      <th>IRQ</th>
      <th>SoftIRQ</th>
      <th>Steal</th>
      <th>Guest</th>
      <th>GuestNice</th>
    </tr>

     <tr>
      <th>CPU {{procstat.CPU.Percent}} %</th>
      <th>{{procstat.CPU.User}}</th>
      <th>{{procstat.CPU.Nice}}</th>
      <th>{{procstat.CPU.System}}</th>
      <th>{{procstat.CPU.Idle}}</th>
      <th>{{procstat.CPU.IOWait}}</th>
      <th>{{procstat.CPU.IRQ}}</th>
      <th>{{procstat.CPU.SoftIRQ}}</th>
      <th>{{procstat.CPU.Steal}}</th>
      <th>{{procstat.CPU.Guest}}</th>
      <th>{{procstat.CPU.GuestNice}}</th>
    </tr>

    {{#each procstat.CPUs:i}}
    <tr>
      <td>CPU{{i}} {{Percent}} %</td>
      <td>{{User}}</td>
      <td>{{Nice}}</td>
      <td>{{System}}</td>
      <td>{{Idle}}</td>
      <td>{{IOWait}}</td>
      <td>{{IRQ}}</td>
      <td>{{SoftIRQ}}</td>
      <td>{{Steal}}</td>
      <td>{{Guest}}</td>
      <td>{{GuestNice}}</td>
    </tr>
    {{/each}}
    </table>`,
    pageMem:`<h3>Mem Info</h3>
    
    <ul>
      {{#each meminfo:varname}}
        <li><b>{{varname}}</b> {{this}}</li>
      {{/each}}
    </ul>`
  }
});

function ajaxGET(url,fGotData){
  var request = new XMLHttpRequest();request.open('GET', url, true);
  request.onload = function() {
    if(this.status>=200 && this.status<400){
      if(fGotData!=undefined){
        try{
          fGotData(JSON.parse(this.response));
          return;
        }catch(e){
          fGotData(this.response);
          return
        }
      }
    }else{}
  };
  request.onerror = function() {};
  request.send();
}

function stringPOST(url,data,resp,respErr){
  var request = new XMLHttpRequest();
  request.open('POST', url, true);
  request.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded; charset=UTF-8');
  request.onreadystatechange = function() {
    console.log("onreadystatechange readystate="+request.readyState+" status="+request.status)
    if (request.readyState == 4) {
      if (request.status == 200) {
        resp(request.responseText);
      }else{
        respErr(request.responseText)
      }
    }
  }
  request.send(data);
}

ractive.on({
  doElfUpload:function(a, target){
      console.log("Doing upload elf binary to "+target)
      const fileInput = document.getElementById('progFileInput');
   
      const files = fileInput.files;
      console.log("have "+files.length+" files")
      
      if (files.length === 0) {
          alert('No files selected.');
          return;
      }
      const formData = new FormData();
        formData.append('program', files[0]);
      

      let pars=new URLSearchParams(window.location.search)
      let xhttp = new XMLHttpRequest();
      xhttp.open('POST', target, true);
      xhttp.onreadystatechange = function() {//Call a function when the state changes.
          if(xhttp.readyState == 4 && xhttp.status == 200) {
              console.log("RECIEVED:"+xhttp.responseText);   
          }
      }
      xhttp.send(formData);
  },


  setenv:function(){
    stringPOST("/env",ractive.get("env"),function(s){
        ractive.set("env",s)
    },function(errmsg){
      ractive.set("envFeedback",errmsg)
    })
  },

  doRestart:function(){
    ajaxGET("/restart",function(a,s){
      console.log(s)
    })
  },
  doReboot:function(){
    ajaxGET("/reboot",function(a,s){
      console.log(s)
    })
  },
  setTab:function(a,pg){
    console.log("uus sivu "+pg)
    ractive.set("page",pg)

    ractive.set("envFeedback","") //hack
    switch(pg){
      case "stdout":
        ajaxGET("/stdout",function(s){
          ractive.set("stdout",s)
        })
      break
      case "stderr":
        ajaxGET("/stderr",function(s){
          ractive.set("stderr",s)
        })
      break
      case "chrashmsg":
        ajaxGET("/chrashmsg",function(s){
          ractive.set("chrashmsg",s)
        })
      break
      case "env":
        ajaxGET("/env",function(s){
          ractive.set("env",s)
        })
      break
      case "kmsg":
        ajaxGET("/kmsg",function(s){
          ractive.set("kmsg",s)
        })
      break
      case "mounts":
        ajaxGET("/mounts",function(s){
          ractive.set("mounts",s)
        })
      break
      case "blockdevices":
        ajaxGET("/blockdevices",function(s){
          ractive.set("blockdevices",s)
        })
        ajaxGET("/filehandleusage",function(s){
          ractive.set("filehandleusage",s)//Text formatted?? vs json?
        })
      break
      case "procstat":
        ajaxGET("/procstat",function(s){
          s.CPU.TotalCpuTime=s.CPU.User + s.CPU.Nice + s.CPU.System + s.CPU.Idle + s.CPU.IOWait + s.CPU.IRQ + s.CPU.SoftIRQ +s.CPU.Steal + s.CPU.System
          s.CPU.Percent = Math.floor(100 * (1.0 - s.CPU.Idle/s.CPU.TotalCpuTime))
          for(let i in s.CPUs){
            s.CPUs[i].TotalCpuTime=s.CPUs[i].User + s.CPUs[i].Nice + s.CPUs[i].System + s.CPUs[i].Idle + s.CPUs[i].IOWait + s.CPUs[i].IRQ + s.CPUs[i].SoftIRQ +s.CPUs[i].Steal + s.CPUs[i].System
            s.CPUs[i].Percent = Math.floor(100 * (1.0 - s.CPUs[i].Idle/s.CPUs[i].TotalCpuTime))
          }
          ractive.set("procstat",s)
        })
      break
      case "mem":
        ajaxGET("/meminfo",function(s){
          ractive.set("meminfo",s)
        })
    }
  }
})

/*
// Handle file upload
document.getElementById('uploadForm').addEventListener('submit', function (e) {
  e.preventDefault();
  const elfFile = document.getElementById('elfFile').files[0];
  const initramfsFile = document.getElementById('initramfsFile').files[0];

  if (elfFile && initramfsFile) {
    alert(`Files uploaded:\n- ELF: ${elfFile.name}\n- Initramfs: ${initramfsFile.name}`);
    // Simulate processing
    ractive.set('stdout', 'Processing uploaded files...');
    setTimeout(() => {
      ractive.set('stdout', 'Files processed successfully!');
    }, 2000);
  } else {
    alert('Please upload both files.');
  }
});*/


/*
// Tabbed interface functionality
function setActiveTab(tabName) {
  // Hide all tab content
  document.querySelectorAll('.tabcontent').forEach(tab => {
    tab.style.display = 'none';
  });

  // Remove active class from all tabs
  document.querySelectorAll('.tablink').forEach(tab => {
    tab.classList.remove('active');
  });

  // Show the selected tab content and mark the button as active
  document.getElementById(tabName).style.display = 'block';
  event.target.classList.add('active');

  // Simulate data updates
  if (tabName === 'cpu') {
    ractive.set('cpuLoad', `CPU Load: ${(Math.random() * 100).toFixed(2)}%`);
  } else if (tabName === 'disk') {
    ractive.set('diskStatus', `Disk Usage: ${(Math.random() * 100).toFixed(2)}%`);
  }
    
}

// Set default tab
setActiveTab('stdout');

*/