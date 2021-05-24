import CommandLineTerminal from '@hashicorp/react-command-line-terminal'
import { useEffect, useState } from 'react'

export default function AnimatedTerminal({
  lines,
  frameLength = 1000,
  paused,
  loop,
}) {
  // Determine the total number of frames
  let totalFrames = 0
  lines.forEach((line) => {
    let frames = line.frames ? line.frames : 1
    if (Array.isArray(line.code)) {
      totalFrames += line.code.length * frames
    } else {
      totalFrames += frames
    }
  })

  // Set up Animation
  const [frame, setFrame] = useState(0)
  useEffect(() => {
    let interval = setInterval(() => {
      if (paused) return
      if (loop) return setFrame((frame) => frame + 1)
      if (frame + 1 < totalFrames) {
        setFrame((frame) => frame + 1)
      }
    }, frameLength)
    return () => clearInterval(interval)
  }, [frame])

  // Reset Frames if our lines change
  useEffect(() => {
    setFrame(0)
  }, [lines])

  const renderedLines = [...lines.slice(0, frame)]

  return <CommandLineTerminal product="packer" lines={renderedLines} />
}
