import React from 'react';
import {sample_proto} from '../dummy/sample_proto'
import { AccessDB } from 'react-indexed-db';
import Modal from 'react-modal';
var fileDownload = require('js-file-download');
export default class ProtoTab extends React.Component {
  state ={
    "textarea":sample_proto,
    "protofiles":{"tf.mock":sample_proto},
    "value": "",
    "fileinput":"",
    modalIsOpen:false
  }
  
  constructor(prop){
    super(prop)
    this.state.protofiles = prop.protofiles || {"tf.mock":sample_proto}
    var z;
    for (var p in prop.protofiles){
      z = p;
    }
    this.state.value = z || "tf.mock"
    this.state.fileinput = z || "tf.mock"
  }
  uploadInput = null
  change=event=>{
    var protofiles = this.state.protofiles
    var textarea = ""
    console.log("event",event.target.value)
    for (var property in protofiles){
      if (property==event.target.value){
        textarea = protofiles[property]
        break;
      }
    }
    this.setState({value:event.target.value,textarea:textarea})
  }
  changeTextInput=event=>{
    this.setState({fileinput:event.target.value})
  }
  changeTextArea=event=>{
    if (this.state.fileinput!=""){
      var protofiles = this.state.protofiles
      protofiles[this.state.fileinput] = event.target.value
      this.setState({protofiles:protofiles,textarea:event.target.value})
    }
  }
  set=()=>{
    var protofiles = this.state.protofiles
    protofiles[this.state.fileinput]=this.state.textarea
    this.setState({protofiles:protofiles})
  }
  delete=()=>{
    var protofiles = this.state.protofiles
    delete protofiles[this.state.value]
    console.log("protofiles",protofiles,this.state.value)
    var z;
    var textarea=""
    for (var p in protofiles){
      z=p;
      textarea=protofiles[z];
      break;
    }
    this.setState({protofiles:protofiles,value:z,textarea:textarea,fileinput:z})
  }
  uploadChangeInput = (ref)=>{
    var file = ref.target.files[0];
    var __this = this
    if (file) {
        var reader = new FileReader();
        
        reader.onload = function (evt) {
            console.log(evt);
            var protofiles= JSON.parse(evt.target.result)
            var first = "nil"
            for (var p in protofiles){
              first = protofiles[p]
              break;
            }
            __this.setState({protofiles:protofiles,modalIsOpen:false,textarea:first});
        };

        reader.onerror = function (evt) {
            console.error("An error ocurred reading the file",evt);
        };

        reader.readAsText(file, "UTF-8");
    }
  }
  saveInput = ()=>{
    fileDownload(JSON.stringify(this.state.protofiles),"protofiles.json")
  }
  closeModal=()=>{
    this.setState({modalIsOpen:false})
  }
  openModal=()=>{
    this.setState({modalIsOpen:true})
  }
  componentDidMount(){
    var __this = this
    var request = window.indexedDB.open("MyDB");
    request.onsuccess = function(event) {
      var db = event.target.result;
      console.log("db",db)
      var objectStore = db.transaction("protofile").objectStore("protofile");
      objectStore.getAll().onsuccess = function(event) {
        var protofiles ={}
        for (var p in event.target.result){
          var obj = event.target.result[p]
          protofiles[obj["name"]]= obj["content"]
        }
        window["protofiles"] = protofiles
        __this.setState({protofiles:protofiles})
      };
    };
    request.onerror = function (evt) {
      console.error("error", evt);
    };
    request.onblocked = function (evt) {
        console.log('blocked', evt);
    };
  }
  render() {
    return (
      <div style={{width:"100%"}}>
        <span><select onChange={this.change} value={this.state.value}>
        {Object.keys(this.state.protofiles).map((value,index)=>{
            return <option value={value} key={value}>{value}</option>
          })}
        </select></span>
        <span><button onClick={this.delete}>Delete</button></span>
        <span><button onClick={this.openModal}>Upload</button></span>
        <div>Protofile</div>
        <textarea value={this.state.textarea} onChange={this.changeTextArea} style={{resize:"none",width:"100%",height:"100px"}}></textarea>
        <div>filename:</div>
        <input onChange={this.changeTextInput} value={this.state.fileinput}/>
        <AccessDB objectStore="protofile" version={1}>
          {({ add }) => {
            const handleClick = () => {
              add({ name: this.state.fileinput, content: this.state.textarea }).then(
                event => {
                  console.log('ID Generated: ', this.state.fileinput);
                },
                error => {
                  console.log(error);
                }
              );
              var protofiles = this.state.protofiles
              protofiles[this.state.fileinput]=this.state.textarea
              this.setState({protofiles:protofiles})
            };
            return <button onClick={handleClick}>Set</button>;
          }}
        </AccessDB>
        <Modal
          isOpen={this.state.modalIsOpen}
          onRequestClose={this.closeModal}
          contentLabel="Upload proto"
        >
          <input  onChange={this.uploadChangeInput} ref={(ref) => { this.uploadInput = ref; }}  type="file" />
          <button onClick={this.saveInput}>Download</button>
          <br></br>
          <button onClick={this.closeModal}>close</button>
          <div></div>
        </Modal>
      </div>
    )
  }
}