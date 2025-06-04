import { useEffect } from "react"

import { useState } from "react"

 
export function Welcome() {
  const DOGENET_ADDRESSES = ["127.0.0.1:8085", "127.0.0.1:8086"]

  const [dogenetAddresses, setDogenetAddresses] = useState(DOGENET_ADDRESSES)
  const [nodes, setNodes] = useState([])

  useEffect(() => {
    const fetchNodes = async (dogenetAddress: string) => {
      const response = await fetch(`http://${dogenetAddress}/nodes`)
      const data = await response.json()
      setNodes(data.nodes)
    }  

    for (const address of DOGENET_ADDRESSES) {
      fetchNodes(address)
    }
  }, [])

  return (
    <main className="flex items-center justify-center pt-16 pb-4">
      <div className="flex flex-col gap-4">
        {dogenetAddresses.map((address) => (
          <div key={address}>{address}</div>
        ))}
      </div>
    </main>
  );
}
 
  