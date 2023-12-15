import React from 'react';
import Editor from 'react-simple-code-editor';
import { highlight, languages } from 'prismjs/components/prism-core';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-javascript';
import "prismjs/themes/prism.css";
const code = `function add(a, b) {
  return a + b;
}
`
class CodeEditor extends React.Component {
  state = { code };
  constructor(props){
    super(props)
  }
  onValueChange=function(value){
    
  }
  render() {
    return (
      <div >
      <Editor
        className="box"
        value={this.state.code}
        onValueChange={code => this.setState({ code })}
        highlight={code => highlight(code, languages.js)}
        padding={10}
        style={{
          fontFamily: '"Fira code", "Fira Mono", monospace',
          fontSize: 12,
        }}
      />
      </div>
    );
  }
}

export default CodeEditor