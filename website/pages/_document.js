import Document, { Html, Head, Main, NextScript } from 'next/document'
import HashiHead from '@hashicorp/react-head'

export default class MyDocument extends Document {
  render() {
    return (
      <Html>
        <Head>
          <HashiHead />
        </Head>
        <body>
          <Main />
          <NextScript />
          <script
            noModule
            dangerouslySetInnerHTML={{
              __html: `window.MSInputMethodContext && document.documentMode && document.write('<script src="/ie-custom-properties.js"><\\x2fscript>');`,
            }}
          />
        </body>
      </Html>
    )
  }
}
