import React from "react";
import logo from "./logo.svg";
import Button from "@mui/material/Button";
import rtc from "./rtc";

function App() {
  const videoEl = React.createRef<HTMLVideoElement>()
  const handleClick = async () => {
    rtc.initTrack(videoEl.current!)
    rtc.start();
    const offer = await rtc.createOffer()
    const response = await fetch("http://localhost:3001/stream/webrtc", {method: "POST", body: JSON.stringify(offer)})
    const answer = await response.json()
    console.log(answer)
    rtc.acceptOffer(answer)
    // rtc.sendSdpToSignaling(offer.sdp)
  };

  return (
    <div className="App" style={{ padding: 10 }}>
      <Button variant="contained" onClick={handleClick}>
        Connect
      </Button>

      <div style={{margin: 16}}>
        <video style={{background: "#eee"}} width={640} ref={videoEl} controls></video>
      </div>
    </div>
  );
}

export default App;
