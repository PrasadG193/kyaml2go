const URL = "http://localhost:8080/v1/convert?method="

let go = document.getElementById("goGenerator")


window.generatorCall=function (api){
  let yamlData  = document.getElementById("yamlGenerator").value
  document.getElementById('yamlGenerator').style.border = "1px solid #ced4da"
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
      document.getElementById('yamlGenerator').style.border = "1px solid red"
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
