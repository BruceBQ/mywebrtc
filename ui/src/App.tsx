import React from "react";
import logo from "./logo.svg";
import Button from "@mui/material/Button";
import rtc from "./rtc";

function App() {
  const handleClick = () => {
    rtc.start();
  };

  return (
    <div className="App" style={{ padding: 10 }}>
      <Button variant="contained" onClick={handleClick}>
        Connect
      </Button>
    </div>
  );
}

export default App;
