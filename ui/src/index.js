import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import { IndexedDB } from 'react-indexed-db';
export const DBConfig = {
  name: 'MyDB',
  version: 1,
  objectStoresMeta: [
    {
      store: 'protofile',
      storeConfig: { keyPath: 'id', autoIncrement: true },
      storeSchema: [
        { name: 'name', keypath: 'name', options: { unique: false } },
        { name: 'content', keypath: 'content', options: { unique: false } }
      ]
    }
  ]
};
//initDB(DBConfig)
ReactDOM.render(
  <React.StrictMode>
     <IndexedDB
      name="MyDB"
      version={1}
      objectStoresMeta={[
        {
          store: 'protofile',
          storeConfig: { keyPath: 'id', autoIncrement: true },
          storeSchema: [
            { name: 'name', keypath: 'name', options: { unique: false } },
            { name: 'content', keypath: 'content', options: { unique: false } }
          ]
        }
      ]}>
      <App />
    </IndexedDB>
    
  </React.StrictMode>,
  document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
