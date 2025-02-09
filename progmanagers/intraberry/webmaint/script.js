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
    pageDisk:"<h3>Disk Status</h3><p>{{diskStatus}}</p>"
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