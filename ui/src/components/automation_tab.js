import React from 'react';
import Editor from 'react-simple-code-editor';
import { highlight, languages } from 'prismjs/components/prism-core';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-javascript';
import "prismjs/themes/prism.css";
import { Grid } from '@material-ui/core';
import Pbf from 'pbf/index'
import Modal from 'react-modal';
import {convertUint8ArrayToBinaryString,convertBinaryStringToUint8Array} from '../utilities/marshaller'
import {deepEqual} from '../utilities/deepcompare'
var fileDownload = require('js-file-download');

var url_string = window.location.href
var url = new URL(url_string);
var domain = url.searchParams.get("domain");
if (domain==null){
   domain = window["location"].href.split("/index.html")[0].split("//")[1]
}

var INDEX = 0
export default class AutomationTab extends React.Component {
  state = {
    mockCommands:"",
    code:`REGISTERFUNC['request']=function(msg){
  var z = {'data':'hi'}
  var pbf = new PbfObj();
  proto['tf.mock']['EchoRequest'].write(z, pbf);
  var buf = pbf.finish();
  return buf;
};
REGISTERFUNC['command']=function(msg){
  return 'tf.mock.echo' 
};
REGISTERFUNC['request_marshalling']=function(msg){
  var z = proto['tf.mock']['EchoRequest'].read(new PbfObj(msg))
  return JSON.stringify(z) 
};
REGISTERFUNC['response_marshalling']=function(msg){
  var z = proto['tf.mock']['EchoResponse'].read(new PbfObj(msg))
  var z2 = {'data':'hello'}
  ASSERT_EQUAL(z.data,z2.data,'is hello')
  ASSERT_NOT_EQUAL(z.data,'james','should not be james')
  return JSON.stringify(z) 
};
`,
    mockCodes:[],
    modalIsOpen: false,
    modalContent: "",
    number: 1,
    spinning: false,
    reportUid: ""
  }
  socket_mock={}
  UID =""

  changeMockCommand=event=>{
    this.setState({mockCommands:event.target.value})
  }
  changeSetMockCommandByIndex=(index,text)=>{
    var mockCodes = this.state.mockCodes
    if (mockCodes.length>index){
      mockCodes[index]["command"]=text
      this.setState({mockCodes:mockCodes})
    }
  }
  changeNumber=event=>{
    this.setState({number:event.target.value})
  }
  startWS=()=>{
    this.initWSMock(this.state.mockCommands)
  }
  initWSMock=(targets)=>{
    var __this = this
    var PbfObj = Pbf
    this.socket_mock = new WebSocket("ws:"+domain+"/ws/set_mock");
    this.socket_mock.onopen = function() {
      __this.socket_mock.send(targets)
        console.log("socket_mock open")
    }
    this.socket_mock.onmessage = function (e) {
      var j = JSON.parse(e.data)
      var proto = window["proto"] 
      var protofiles_map = window["protofiles_map"]
      var REGISTERFUNC ={}
      for (var i=0;i<__this.state.mockCodes.length;i++){
        var c = __this.state.mockCodes[i]["command"]
        eval(__this.state.mockCodes[i]["code"])
        for (let l=0;l<protofiles_map.length;l++){
          if (protofiles_map[l]["command"]==c){
            REGISTERFUNC[c+"_modify_req"]=function(msg){
              var b = proto[protofiles_map[l]["proto_name"]][protofiles_map[l]["request"]].read(new PbfObj(msg))
              var payload = new TextEncoder().encode(JSON.stringify(b))
              return payload
            }
            REGISTERFUNC[c+"_modify_res"]=function(msg){
              var b = proto[protofiles_map[l]["proto_name"]][protofiles_map[l]["response"]].read(new PbfObj(msg))
              var payload = new TextEncoder().encode(JSON.stringify(b))
              return payload
            }
          }
        } 
      }
      var buf = convertBinaryStringToUint8Array(j["payload"])
      var res = REGISTERFUNC[j["fn"]](buf)
      __this.socket_mock.send(convertUint8ArrayToBinaryString(res))
    }
    this.socket_mock.onclose = function () {
      console.log("socket_mock close")
    }
  }
  mockSample=()=>{
    var mock_split = this.state.mockCommands.split(",")
    var mockCodes = []
    if (typeof window["protofiles_map"]!="undefined"){
    mock_split.forEach(function(v){
        for (var i=0;i<window["protofiles_map"].length;i++){
          if (v==window["protofiles_map"][i]["command"]){
            var proto_name =  window["protofiles_map"][i]["proto_name"];
            var command = window["protofiles_map"][i]["command"];
            var request = window["protofiles_map"][i]["request"];
            var response = window["protofiles_map"][i]["response"];
            var j = `REGISTERFUNC['`+command+`']=function(msg){ 
  var b = proto['`+proto_name+`']['`+request+`'].read(new PbfObj(msg))
  var z = {
   data: "hello"
  }
  var pbf = new PbfObj();
  proto['`+proto_name+`']['`+response+`'].write(z, pbf);
  var buf = pbf.finish();
  return buf
};
`
            mockCodes.push({"command":command,"code":j})
            break;
          }
        }
    })
    this.setState({mockCodes:mockCodes})
    }
  }
  closeModal=()=>{
    this.setState({modalIsOpen:false,modalContent:""})
  }
  async run(__this){
    __this.socket_call = new WebSocket("ws:"+domain+"/ws/rpc");
    this.setState({reportUid:""})
    var PbfObj = Pbf;
    __this.socket_call.onmessage = function (e) {
      var j = JSON.parse(e.data)
      var proto = window["proto"]
      if (typeof j["index"]!="undefined"){ //WsCallProtocol
        INDEX = j["index"]
        var REGISTERFUNC ={
          "get_uid":function(msg){
            __this.UID = msg
            __this.setState({reportUid:msg})
            return msg
          }
        }
        
        var HOSTCALL = {
          "assert_pass":function (msg){
            var j = {
              "fn":"assert_pass",
              "binding":__this.UID,
              "payload":msg
            }
            __this.socket_call.send(JSON.stringify(j))
          },
          "assert_fail":function (msg){
            var j = {
              "fn":"assert_fail",
              "binding":__this.UID,
              "payload":msg
            }
            __this.socket_call.send(JSON.stringify(j))
          },
          "sleep":function (msg){
            var j = {
              "fn":"sleep",
              "binding":"default",
              "payload":msg
            }
            __this.socket_call.send(JSON.stringify(j))
          }
        }
        var ASSERT_EQUAL = function(a,b,msg){
          if (deepEqual(a,b)){
            HOSTCALL["assert_pass"](JSON.stringify(a)+" "+msg)
          }else{
            HOSTCALL["assert_fail"]("left:"+JSON.stringify(a)+" right:"+JSON.stringify(b)+" "+msg)
          }
        }
        var ASSERT_NOT_EQUAL = function(a,b,msg){
          if (deepEqual(a,b)){
            HOSTCALL["assert_fail"](JSON.stringify(a)+" "+msg)
          }else{
            HOSTCALL["assert_pass"]("left:"+JSON.stringify(a)+" right:"+JSON.stringify(b)+" "+msg)
          }
        }
        var SLEEP = function(second){
          var longToByteArray = function(/*long*/long) {
            // we want to represent the input as a 8-bytes array
            var byteArray = [0, 0, 0, 0, 0, 0, 0, 0];
        
            for ( var index = 0; index < byteArray.length; index ++ ) {
                var byte = long & 0xff;
                byteArray [ index ] = byte;
                long = (long - byte) / 256 ;
            }
        
            return new Uint8Array(byteArray);
          };
          var n = longToByteArray(second)
          var m =  convertUint8ArrayToBinaryString(n)
          
          HOSTCALL["sleep"](m)
        }
        eval(__this.state.code)
        switch (j["fn"]){
          case "get_uid":
            var payload = REGISTERFUNC["get_uid"](j["payload"])
            __this.socket_call.send("2")
            break;
          case "command":
            var payload = REGISTERFUNC["command"](j["payload"])
            __this.socket_call.send(payload)
            break;
          case "request":
            var payload = REGISTERFUNC["request"](j["payload"])
            __this.socket_call.send(convertUint8ArrayToBinaryString(payload))
            break;
          case "request_marshalling":
            var payload = REGISTERFUNC["request_marshalling"](convertBinaryStringToUint8Array(j["payload"]))
            __this.socket_call.send(payload)
            break;
          case "response_marshalling":
            var payload = REGISTERFUNC["response_marshalling"](convertBinaryStringToUint8Array(j["payload"]))
            __this.socket_call.send(payload)
            break;
          default:
            break;
        }
      }
    }
    __this.socket_call.onopen =function(){
      console.log("socket_call open")
      __this.socket_call.send(JSON.stringify({"loop":__this.state.number.toString(),"targets":__this.state.mockCommands}))
    }
    
    __this.socket_call.onclose = function () {
      console.log("socket_mock close")
    }
  }
  uploadCallRPC = (ref)=>{
    var file = ref.target.files[0];
    var __this = this
    if (file) {
        var reader = new FileReader();
        reader.onload = function (evt) {
            console.log(evt);
            __this.setState({code:evt.target.result});
        };
        reader.onerror = function (evt) {
            console.error("An error ocurred reading the file",evt);
        };
        reader.readAsText(file, "UTF-8");
    }
  }
  uploadSetMock = (ref)=>{
    var file = ref.target.files[0];
    var __this = this
    if (file) {
        var reader = new FileReader();
        
        reader.onload = function (evt) {
          var arr = evt.target.result.split("\n");
          var mockCodes = []
          var obj = {
            "command":"",
            "code":""
          }
          var mockCommands = [];
          for (var i = 0; i < arr.length; i++){
            if (arr[i].includes("// COMMAND:")){
              var matches = arr[i].match(/\[(.*?)\]/);
              if (matches) {
                obj["command"] =  matches[1];
                obj["code"]=""
                mockCommands.push(matches[1])
              }
            }
            // Do something with arr
            else if (arr[i].includes("// END")){
              mockCodes.push(obj)
            }else if (arr[i]!=""){
              obj["code"]+= arr[i]+"\n"
            }
          }
          __this.setState({mockCodes:mockCodes,mockCommands:mockCommands.join(",")});
        };

        reader.onerror = function (evt) {
            console.error("An error ocurred reading the file",evt);
        };
        reader.readAsText(file, "UTF-8");
    }
  }
  saveSetMock =()=>{
    var f = ""
    this.state.mockCodes.forEach(function(z){
      f+="// COMMAND: ["+z.command +"] \n"
      f+=z.code
      f+="\n\n"
    })
    fileDownload(f,"set_mock.js")
  }
  render(){
    var __this = this
    return (
      <div style={{width:"100%"}}>
        Mock commands:
        <input value={this.state.mockCommands} onChange={this.changeMockCommand}/>
        <button onClick={this.startWS}>Start Websocket</button>
        <button onClick={this.mockSample}>Mock Sample</button>
        <button onClick={()=>{this.setState({modalIsOpen:true})}}>Upload</button>
        <Grid container spacing={1}>
          <Grid item xs={6}>
            <div>Call_rpc</div>
            <div style={{borderColor: "black",
                borderWidth:"3px"}}>
            <Editor
              value={this.state.code}
              onValueChange={code => this.setState({ code })}
              highlight={code => highlight(code, languages.js)}
              padding={10}
              style={{
                fontFamily: '"Fira code", "Fira Mono", monospace',
                fontSize: 12,
                border:"1px solid gray"
              }}
            />
            </div>
            number of iteration:
            <input type="number" value={this.state.number} onChange={this.changeNumber}/>
            <button onClick={()=>this.run(__this)}>Run</button>
            {this.state.reportUid!="" && 
            <a href={"http:"+domain+"/report/"+this.state.reportUid}>report</a>
            }
          </Grid>
          <Grid item xs={6}>
            <div>Set_Mock</div>
            {
              this.state.mockCodes.map(function(v,i){
                return (
                  <div key={v["command"]+"_div"}>
                    <input value={v["command"]} onChange={event=>{__this.changeSetMockCommandByIndex(i,event.target.value)}}/>
                    <Editor
                      key={v["command"]}
                      value={v["code"]}
                      onValueChange={code => {
                        var mockCodes = __this.state.mockCodes
                        mockCodes[i]["code"]=code
                        __this.setState({ mockCodes:mockCodes })}
                      }
                      highlight={code => highlight(code, languages.js)}
                      padding={10}
                      style={{
                        fontFamily: '"Fira code", "Fira Mono", monospace',
                        fontSize: 12,
                        border:"1px solid gray"
                      }}
                    />
                  </div>
                )
              })
            }
          </Grid>
        </Grid>
        <Modal
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Example Modal"
        >
          <label>Request</label>
          <input  onChange={this.uploadCallRPC}  type="file" />
          <button onClick={this.saveCallRPC}>Download CallRPC</button>
          <br></br>
          <label>SetMock</label>
          <input  onChange={this.uploadSetMock}  type="file" />
          <button onClick={this.saveSetMock}>Download SetMock</button>
          <br></br>
          <button onClick={this.closeModal}>close</button>

          <div>{this.state.modalContent}</div>
        </Modal>
      </div>
    )
  }
}