//const URL = "http://localhost:8080/v1/convert"
let BASE_URL = "https://us-central1-yaml2go.cloudfunctions.net/kgoclient-gen?method="

let go = document.getElementById("goGenerator")
let codecopied = document.getElementById("codecopied")
let editor = ""

window.generatorCall=function (action){
  URL = formGooleFuncURL(action)
  let yamlData  = document.getElementById("codegen").value
  document.getElementById('codegen').style.border = "1px solid #ced4da"
  yamlData = editor.getValue()
  $.ajax({
    'url' : `${URL}`,
    'type' : 'POST',
    'data' : yamlData,
    'success' : function(data) { 
        document.getElementById("error").style.display="none" 
        document.getElementById("err-span").innerHTML="";     
        go.setValue(data)
    },
    'error' : function(jqXHR, request,error)
    {
      document.getElementById('codegen').style.border = "1px solid red"
      if (jqXHR.status == 400) {
        // empty out the second textarea
        displayError('Invalid yaml format')
      } else {
        displayError('Something went wrong! Please report this to me@prasadg.dev')
      }
    }
  });

}

function formGooleFuncURL(action){
  return BASE_URL+action
}

//Convert
dropDown = document.getElementById("selectaction")
document.getElementById("convert").addEventListener('click', ()=>{
  action = dropDown.value
  if (action != "select"){
    hideError()
    generatorCall(action)
  }
  else{
    displayError("Please select the method.")
  }
})

//Clear YAML
document.getElementById('clearYaml').addEventListener('click',()=>{
  editor.setValue('')
})

//Clear Go
document.getElementById('clearGo').addEventListener('click',()=>{
  go.setValue('')
})


$(document).ready(function(){
    //code here...
    var input = $(".codemirror-textarea")[0];
    var output = $(".codemirror-textarea")[1];
    editor = CodeMirror.fromTextArea(input, {
        mode: "text/x-yaml",
    	lineNumbers : true
    });

    go = CodeMirror.fromTextArea(output, {
    	lineNumbers : true,
        mode: "text/x-go"
    });

    editor.setValue('# Paste your k8s spec yaml here...\n')

    go.setValue('// Go\n')
});

function displayError(err){
  document.getElementById("err-span").innerHTML=err;
  document.getElementById("error").style.display="block"
}

function hideError(){
  document.getElementById("err-span").innerHTML="";
  document.getElementById("error").style.display="none"
}


document.getElementById("copybutton").addEventListener("click", function (){
  // will have to check browser compatibility for this
  navigator.clipboard.writeText(go.getValue())
  codecopied.style.display="inline"
  window.setTimeout(function (){
    codecopied.style.display="none"
  }, 1500)
});