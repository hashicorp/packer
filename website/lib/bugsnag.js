import React from 'react'
import Bugsnag from '@bugsnag/js'
import BugsnagReact from '@bugsnag/plugin-react'

const apiKey =
  typeof window === 'undefined'
    ? 'b6c57b27a37e531a5de94f065dd98bc0'
    : 'de0b822b269aa57b620efd8927e03744'

if (!Bugsnag._client) {
  Bugsnag.start({
    apiKey,
    plugins: [new BugsnagReact(React)],
    otherOptions: { releaseStage: process.env.NODE_ENV || 'development' },
  })
}

export default Bugsnag
