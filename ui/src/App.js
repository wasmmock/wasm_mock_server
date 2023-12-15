import React,{useState} from 'react';
import {
  BrowserRouter as Router,
  Routes, //replaces "Switch" used till v5
  Route,
} from "react-router-dom";
import './App.css';
import ProtoTab from './components/proto_tab';
import RequestTab from './components/request_tab';
import AutomationTab from './components/automation_tab';
import LoginPage from './navigations/login'
import Grid from '@material-ui/core/Grid';
import Paper from '@material-ui/core/Paper';
import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles((theme) => ({
  root: {
    flexGrow: 1,
    padding: theme.spacing(2),
  },
  paper: {
    padding: theme.spacing(3),
    textAlign: 'center',
    color: theme.palette.text.secondary,
  }
}));
function InnerApp(){
  const classes = useStyles();
  return (
     <div className={classes.root}>
    
      <a href="/">home</a>
      <Paper>
      <Grid container spacing={1}>
        
        <Grid item xs={3}>
          <Grid item xs={12}>
          <Paper className={classes.paper}>
          <ProtoTab/>
          </Paper>
          </Grid>
          <Grid item xs={12}>
          <Paper className={classes.paper}>
            <RequestTab/>
          </Paper>
          </Grid>
        </Grid>
        <Grid item xs={9}>
          <Paper className={classes.paper}>
            <AutomationTab/>
          </Paper>
        </Grid>
      </Grid>
      </Paper>
    </div>          
  )
}
// export default function App() {
//   const classes = useStyles();
  
//   return (
//     <div className="App">
//      <Router>
//         <div className="container">
//         <Routes>
//             <Route path="/" element={<InnerApp />} />
//             <Route path="/login" element={LoginPage(classes)} />
//             {/* <Route path="/dashboard" element={<Dashboard />} /> */}
//         </Routes>
//         </div>
//       </Router>
//    </div>
//   );
  
// }


export default function App() {
  
  const classes = useStyles();
  
  return (
    <div className={classes.root}>
      <a href="/">home</a>
      <Paper>
      <Grid container spacing={1}>
        
        <Grid item xs={3}>
          <Grid item xs={12}>
          <Paper className={classes.paper}>
          <ProtoTab/>
          </Paper>
          </Grid>
          <Grid item xs={12}>
          <Paper className={classes.paper}>
            <RequestTab/>
          </Paper>
          </Grid>
        </Grid>
        <Grid item xs={9}>
          <Paper className={classes.paper}>
            <AutomationTab/>
          </Paper>
        </Grid>
      </Grid>
      </Paper>
    </div>          
  );
  
}
