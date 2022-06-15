import React from "react";
import logo from "./logo.svg";
import Button from "@mui/material/Button";
import rtc from "./rtc";

function App() {
  const handleClick = async () => {
    rtc.start();
    const offer = await rtc.createOffer()
    rtc.sendSdpToSignaling(offer.sdp)
  };

  return (
    <div className="App" style={{ padding: 10 }}>
      <Button variant="contained" onClick={handleClick}>
        Connect
      </Button>

      <div style={{margin: 16}}>
        <video style={{background: "#eee"}} width={640}></video>
      </div>
    </div>
  );
}

export default App;
