import React from 'react';
import GoogleLogin from 'react-google-login';
const responseGoogleSuccess = (response) => {
  console.log(response);
}
const responseGoogleFailure = (response) => {
  console.error(response);
}
export default class LoginTab extends React.Component {
  state ={
  }
  
  constructor(prop){
    super(prop)
  }
  componentDidMount(){
    const script = document.createElement("script");

    script.src = "https://accounts.google.com/gsi/client";
    script.async = true;

    document.body.appendChild(script);
  }
  
  render() {
    
    return (
      <div style={{width:"100%"}}>
       <div id="g_id_onload"
         data-client_id="991783899494-s7945o047o6pmn2c0hgc9dmm47s5gm6m.apps.googleusercontent.com"
         data-login_uri="http://localhost:20825/login/2/authorization/google"
         data-auto_prompt="false">
      </div>
      <div class="g_id_signin"
         data-type="standard"
         data-size="large"
         data-theme="outline"
         data-text="sign_in_with"
         data-shape="rectangular"
         data-logo_alignment="left">
      </div>
      </div>
    )
  }
}