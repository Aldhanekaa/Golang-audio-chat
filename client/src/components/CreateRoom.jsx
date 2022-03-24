import React, { useRef, useState } from 'react';

const CreateRoom = (props) => {
  const [name, setName] = useState();
  const nameInputRef = useRef();

  const create = async (e) => {
    e.preventDefault();

    const resp = await fetch('http://localhost:8000/create');
    const { room_id } = await resp.json();

    props.history.push(`/room/${room_id}`);
  };

  return (
    <div>
      <input type="text" name="name" ref={nameInputRef} />
      <button
        onClick={(e) => {
          setName(nameInputRef.current.value);
        }}
      >
        Save Name
      </button>

      <br />
      <button onClick={create}>Create Room</button>
    </div>
  );
};

export default CreateRoom;
