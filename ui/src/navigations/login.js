import Grid from '@material-ui/core/Grid';
import Paper from '@material-ui/core/Paper';
import LoginTab from '.././components/login_tab';

export default function LoginPage(classes){
  return (
     <div className={classes.root}>
    
      <a href="/">home</a>
      <Paper>
      <Grid container spacing={1}>
        
        <Grid item xs={3}>
          <Grid item xs={12}>
          <Paper className={classes.paper}>
            <h2>HH</h2>
            <LoginTab/>
          </Paper>
          </Grid>
          <Grid item xs={12}>
          <Paper className={classes.paper}>
           
          </Paper>
          </Grid>
        </Grid>
        <Grid item xs={9}>
          <Paper className={classes.paper}>
          </Paper>
        </Grid>
      </Grid>
      </Paper>
    </div>          
  )
}