const URL = "http://192.168.1.179:8080/v1/convert?method="
// const data = ``

let go = document.getElementById("goGenerator")


window.generatorCall=function (api){
  let yamlData  = document.getElementById("yamlGenerator").value
  console.log(yamlData)
  $.ajax({
    'url' : `${URL}${api}`,
    'type' : 'POST',
    'data' : yamlData,
    'success' : function(data) { 
        go.value = data
    },
    'error' : function(request,error)
    { 
        alert('Something is fishy')      
    }
  });

}

//Create Function
document.getElementById("createGO").addEventListener('click', ()=>{
  generatorCall('create')
})

//Update Function
document.getElementById("updateGO").addEventListener('click', ()=>{
  generatorCall('update')
})

//Delete Function
document.getElementById("deleteGO").addEventListener('click', ()=>{
    generatorCall('delete')
})

//Clear YAML
document.getElementById('clearYaml').addEventListener('click',()=>{
  document.getElementById("yamlGenerator").value = ''
})

//Clear Go

document.getElementById('clearGo').addEventListener('click',()=>{
  document.getElementById("goGenerator").value = ''
})