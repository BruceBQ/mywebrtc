import signaling from "./signaling";

class RTC {
  localConnection: RTCPeerConnection;
  localTracks: MediaStreamTrack[] = [];
  constructor() {
    this.localConnection = this.createLocalPeerConnection();
  }

  createLocalPeerConnection = () => {
    const pc = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    });

    pc.onconnectionstatechange = (e) => {
      console.log("onconnectionstatechange", e);
    };
    pc.ondatachannel = (e) => {
      console.log("ondatachannel", e);
    };
    pc.onicecandidate = (e) => {
      console.log("onicecandidate", e);
    };
    pc.onicecandidateerror = (e) => {
      console.log("onicecandidateerror", e);
    };
    pc.oniceconnectionstatechange = (e) => {
      console.log("oniceconnectionstatechange", e);
    };
    pc.onicegatheringstatechange = (e) => {
      console.log("onicegatheringstatechange", e);
    };
    pc.onnegotiationneeded = (e) => {
      console.log("onnegotiationneeded", e);
    };
    pc.onsignalingstatechange = (e) => {
      console.log("onsignalingstatechange", e);
    };
    pc.ontrack = (e) => {
      console.log("ontrack", e);
    };

    return pc;
  };

  createLocalTracks = () => {
    const media = new MediaStream();
    return media;
  };

  start = () => {
    const stream = this.createLocalTracks();
    const tracks = stream.getVideoTracks();
    for (let i = 0; i < tracks.length; i++) {
      console.log("Track", tracks[i])
      this.localConnection.addTrack(tracks[i]);
    }
    signaling.connect()
  };

  stop = (closeConnection: Boolean) => {};
}

const rtc = new RTC();

export default rtc;
