export function parse_commands(txt,name){
  var lines = txt.split("\n");
  var start = false;
  var captured =[]

  lines.forEach(function(v,i){
    
    if (start){
      if (v.includes("}")){
        start = false;
      }
      var k = v.replace(/\/\//g, '').replace(/ /g,'');
      try{
        var o = {
          "proto_name":name,
          "command":k.split("(")[0],
          "request":k.split("(")[1].split(",")[0],
          "response":k.split(",")[1].split(")")[0]
        }
        captured.push(o)
      }catch(err){

      }
      
    }
    if (v.includes("commands")){
      start = true;
    }
    
  })
  return captured
}