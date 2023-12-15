import React from 'react';
import { AccessDB } from 'react-indexed-db';
import {parse_commands} from '../utilities/parse_command'
import Pbf from 'pbf/index'
import Grid from '@material-ui/core/Grid';
import Editor from 'react-simple-code-editor';
import { highlight, languages } from 'prismjs/components/prism-core';
import 'prismjs/components/prism-json';
import "prismjs/themes/prism.css";
import Modal from 'react-modal';
import {convertUint8ArrayToBinaryString,convertBinaryStringToUint8Array,compileProtofiles} from '../utilities/marshaller'

function getParameterByName(name) {
  name = name.replace(/[\[]/, "\\[").replace(/[\]]/, "\\]");
  var regex = new RegExp("[\\?&]" + name + "=([^&#]*)"),
      results = regex.exec(window.location.search);
  return results === null ? "" : decodeURIComponent(results[1].replace(/\+/g, " "));
}
var PbfObj = Pbf  
var url_string = window.location.href
var url = new URL(url_string);
var domain = url.searchParams.get("domain");
if (domain==null){
   //domain = window["location"].href.split(":")[1]
   domain = window["location"].href.split("/index.html")[0].split("//")[1]
}
export default class RequestTab extends React.Component {
  state = {
    textarea: `{"data":"hi"}`,
    command:"",
    request:"",
    response:"",
    protofile_name:"",
    protofiles_map:[],
    value:"",
    response_json:"",
    header:"",
    modalIsOpen:false
  }
  proto ={}
  
  change=event=>{
    var protofiles_map = this.state.protofiles_map
    var command = "";
    var request = "";
    var response = "";
    var protofile_name =""
    for (var i=0;i<protofiles_map.length;i++){
      if (protofiles_map[i]["command"]==event.target.value){
        if (getParameterByName("routing")!=""){
          command = event.target.value + "?"+ getParameterByName("routing")
        }else{
          command = event.target.value
        }
        request = protofiles_map[i]["request"]
        response = protofiles_map[i]["response"]
        protofile_name = protofiles_map[i]["proto_name"]
        break;
      }
    }
    this.setState({value:event.target.value,command:command,request:request,response:response,protofile_name:protofile_name})
  }
  changeTextArea=code=>{
    this.setState({textarea:code})
  }
  changeCommand=event=>{
    this.setState({command:event.target.value})
  }
  changeRequest=event=>{
    this.setState({request:event.target.value})
  }
  changeResponse=event=>{
    this.setState({response:event.target.value})
  }

  send=event=>{
    var proto = this.proto;
    var buf;
    if (getParameterByName("json")=="true"){
      buf = convertBinaryStringToUint8Array(this.state.textarea)
    }else{
      var t = JSON.parse(this.state.textarea)
      var pbf = new PbfObj();
      proto[this.state.protofile_name][this.state.request].write(t, pbf);
      buf = pbf.finish();
    }
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "http://"+domain+"/call/raw_rpc?command="+this.state.command, true);
    xhr.responseType = "arraybuffer";
    var __this= this
    xhr.onload = function (oEvent) {
      var arrayBuffer = xhr.response; // Note: not oReq.responseText
      if (arrayBuffer) {
        var byteArray = new Uint8Array(arrayBuffer);
        var res = proto[__this.state.protofile_name][__this.state.response].read(new Pbf(byteArray))
        __this.setState({response_json:JSON.stringify(res),header:xhr.getAllResponseHeaders()})
      }
    }
    xhr.send(buf);
  }
  closeModal=()=>{
    this.setState({modalIsOpen:false})
  }
  expand=()=>{
    this.setState({modalIsOpen:true})
  }
  componentDidMount(){
    var j = highlight(`{"data":"hi"}`, languages.json)
    console.log("j",j,this.state.response_json)
    this.setState({command:""})
  }
  render(){
    return (
      <div style={{width:"100%"}}>
        <AccessDB objectStore="protofile" version={1}>
        {({ getAll }) => {
          const handleClick = () => {
            getAll().then(
              protofileFromDB => {
                var protofiles_map=[]
                protofileFromDB.map(function(v,i){
                  parse_commands(v["content"],v["name"]).forEach(function(v){
                    var found = false
                    for (var i =0;i<protofiles_map.length;i++){
                      if (protofiles_map[i]["command"]==v["command"]){
                        found = true
                        break;
                      }
                    }
                    if (!found){
                      protofiles_map.push(v)
                    }
                  })
                });
                window["protofiles_map"]= protofiles_map
                var proto = compileProtofiles(protofileFromDB)
                window["proto"] = proto
                this.proto = proto
                this.setState({protofiles_map:protofiles_map})
              },
              error => {
                console.log(error);
              }
            );
            
          };
          return (<button onClick={handleClick}>Refresh</button>
          )
        }}
        </AccessDB>
        <div>Suggested commands:<select id="suggested" onChange={this.change} value={this.state.value} >
            {
              this.state.protofiles_map.map(function(value,i){
                return (<option value={value.command} key={i}>{value.command}</option>)
              })
            }
            </select></div>
        <table style={{width:"100%"}}>
          <tbody style={{width:"100%"}}>
            <tr>
              <td style={{width:"33%"}}>command:<input id="command" value={this.state.command} onChange={this.changeCommand} style={{width:"100%"}}/></td>
              <td style={{width:"33%"}}>request:<input id="request_struct" value={this.state.request} onChange={this.changeRequest} style={{width:"100%"}} /></td>
              <td style={{width:"33%"}}>response:<input id="response_struct" value={this.state.response} onChange={this.changeResponse} style={{width:"100%"}}/></td>
            </tr>
          </tbody>
        </table>
        <Grid container spacing={1}>
        <Grid item xs={10}>
        <div style={{height:"20px"}}>Request Json</div>
        <Editor
          className="box"
          value={this.state.textarea}
          onValueChange={code=>this.changeTextArea(code)}
          highlight={code => highlight(code, languages.json)}
          padding={10}
          style={{
            fontFamily: '"Fira code", "Fira Mono", monospace',
            fontSize: 12,
            width:"100%",height:"100px",
            border:"1px solid gray"
          }}
        />
        </Grid>
        <Grid item xs={2}>
          <div style={{height:"20px"}}> </div>
          <button style={{height:"80%"}} onClick={this.send}>Send</button>
        </Grid>
        </Grid>
        <div>Response</div>
        <div>{this.state.response_json}
        {
          this.state.response_json.length>50 && <button onClick={this.expand}>Expand</button>
        }
        </div>
        <div>Header:{this.state.header}</div>
        <Modal
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Example Modal"
        >
         
          <button onClick={this.closeModal}>close</button>
          <div><pre>{highlight(this.state.response_json, languages.json)}</pre></div>
        </Modal>
      </div>
    )
  }
}
function str2ab(str) {
  var buf = new ArrayBuffer(str.length * 2); // 2 bytes for each char
  var bufView = new Uint16Array(buf);
  for (var i = 0, strLen = str.length; i < strLen; i++) {
    bufView[i] = str.charCodeAt(i);
  }
  return buf;
}