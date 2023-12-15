import schema from 'protocol-buffers-schema'
import {Pbf} from 'pbf'
import compile from 'pbf/compile'
export function convertUint8ArrayToBinaryString(u8Array) {
	var i, len = u8Array.length, b_str = "";
	for (i=0; i<len; i++) {
		b_str += String.fromCharCode(u8Array[i]);
	}
	return b_str;
}
export function convertBinaryStringToUint8Array(bStr) {
	var i, len = bStr.length, u8_array = new Uint8Array(len);
	for (var i = 0; i < len; i++) {
		u8_array[i] = bStr.charCodeAt(i);
	}
	return u8_array;
}
export function compileProtofiles(protofiles){
  var proto ={}
  protofiles.forEach(function(v){
    var sch = schema.parse(v["content"])
    proto[v["name"]]=compile(sch)
  })
  return proto
}