import React, { useEffect, useRef, useState } from 'react';

const Room = (props) => {
  const userVideo = useRef();
  const userStream = useRef();
  const partnerVideo = useRef();
  const peerRef = useRef();
  const webSocketRef = useRef();
  const [participantId, setParticipantId] = useState();

  const openCamera = async () => {
    const allDevices = await navigator.mediaDevices.enumerateDevices();
    const cameras = allDevices.filter((device) => device.kind == 'videoinput');
    console.log(cameras);

    const constraints = {
      audio: true,
      video: {
        deviceId: cameras[0].deviceId,
      },
    };

    try {
      return await navigator.mediaDevices.getUserMedia(constraints);
    } catch (err) {
      console.log(err);
    }
  };

  useEffect(() => {
    if (!userStream.current && !userVideo.current.srcObject) {
      openCamera().then((stream) => {
        console.log('stream', stream);
        userVideo.current.srcObject = stream;
        userStream.current = stream;

        webSocketRef.current = new WebSocket(
          `${
            import.meta.env.PROD
              ? 'wss://golang-webchat-server.sg.aldhaneka.me'
              : 'ws://localhost:8000'
          }/join?roomID=${props.match.params.roomID}`
        );

        webSocketRef.current.addEventListener('close', (event) => {
          console.log('CLOSE! ', event);
          webSocketRef.current.send(JSON.stringify({ message: 'woee' }));
        });

        webSocketRef.current.addEventListener('open', () => {
          webSocketRef.current.send(JSON.stringify({ ask: true }));
        });

        webSocketRef.current.addEventListener('message', async (e) => {
          console.log('on message ', e);
          const message = JSON.parse(e.data);

          // when new user join
          if (message.join && message.participantId) {
            callUser();
            return;
          }

          // get current participantId
          if (message.participantId) {
            setParticipantId(message.participantId);
            return;
          }

          if (message.offer) {
            handleOffer(message.offer);
          }

          if (message.answer) {
            console.log('Receiving Answer');
            peerRef.current.setRemoteDescription(
              new RTCSessionDescription(message.answer)
            );
          }

          if (message.iceCandidate) {
            console.log('Receiving and Adding ICE Candidate');
            try {
              console.log('PEER REF: ', peerRef.current);

              // add other peer to current user
              await peerRef.current.addIceCandidate(message.iceCandidate);
            } catch (err) {
              console.log(peerRef.current);
              console.log('Error Receiving ICE Candidate', err);
              // alert(err);
            }
          }
        });
      });
    }
  });

  useEffect(() => {
    console.log('hey!', participantId);
    if (participantId) {
      createTracksId();
      webSocketRef.current.send(
        JSON.stringify({
          join: true,
          participantId: participantId,
        })
      );
    }
  }, [participantId]);

  const createTracksId = () => {
    userStream.current.getTracks().forEach(async (track) => {
      console.log('track: ', track);

      // peerRef.current.addTrack(track, userStream.current);
    });
  };

  const handleOffer = async (offer) => {
    console.log('Received Offer, Creating Answer');

    if (!peerRef.current) {
      peerRef.current = createPeer();
    }

    await peerRef.current.setRemoteDescription(
      new RTCSessionDescription(offer)
    );
    console.log('Send Track');

    userStream.current.getTracks().forEach(async (track) => {
      console.log('track: ', track);

      peerRef.current.addTrack(track, userStream.current);
    });

    const answer = await peerRef.current.createAnswer();
    await peerRef.current.setLocalDescription(answer);

    console.log('Send Answer');
    webSocketRef.current.send(
      JSON.stringify({ answer: peerRef.current.localDescription })
    );
  };

  const callUser = () => {
    console.log('Calling Other User');
    peerRef.current = createPeer();

    userStream.current.getTracks().forEach((track) => {
      console.log('track: ', track);

      peerRef.current.addTrack(track, userStream.current);
    });
  };

  const createPeer = () => {
    console.log('Creating Peer Connection');
    const peer = new RTCPeerConnection({
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        {
          urls: 'turn:openrelay.metered.ca:443?transport=tcp',
          username: 'openrelayproject',
          credential: 'openrelayproject',
        },
      ],
    });

    peer.onnegotiationneeded = handleNegotiationNeeded;
    peer.onicecandidate = handleIceCandidateEvent;
    peer.ontrack = handleTrackEvent;
    // peer.onsignalingstatechange
    console.log('PEER: ', peer);
    return peer;
  };

  const handleNegotiationNeeded = async () => {
    console.log('Creating Offer');

    try {
      const myOffer = await peerRef.current.createOffer();
      await peerRef.current.setLocalDescription(myOffer);

      webSocketRef.current.send(
        JSON.stringify({ offer: peerRef.current.localDescription })
      );
    } catch (err) {}
  };

  const handleIceCandidateEvent = (e) => {
    console.log('Found Ice Candidate');
    if (e.candidate) {
      console.log(e.candidate);
      webSocketRef.current.send(JSON.stringify({ iceCandidate: e.candidate }));
    }
  };

  const handleTrackEvent = (e) => {
    console.log('Received Tracks');
    console.log('Stream: ', e);

    partnerVideo.current.srcObject = e.streams[0];
    console.log(partnerVideo.current.srcObject);
  };

  return (
    <div>
      <video autoPlay controls={true} ref={userVideo}></video>
      <video autoPlay controls={true} ref={partnerVideo}></video>
    </div>
  );
};

export default Room;
