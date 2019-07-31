// Constants
import {FLUX_RESPONSE_BYTES_LIMIT} from 'src/shared/constants'

// Types
import {CancelBox} from 'src/types/promises'
import {File, Query, CancellationError} from 'src/types'

export interface RunQueryResult {
  csv: string
  didTruncate: boolean
  bytesRead: number
}

export const runQuery = (
  orgID: string,
  query: string,
  extern?: File
): CancelBox<RunQueryResult> => {
  const url = `/api/v2/query?${new URLSearchParams({orgID})}`

  const headers = {
    'Content-Type': 'application/json',
    'Accept-Encoding': 'gzip',
  }

  const body: Query = {
    query,
    extern,
    dialect: {annotations: ['group', 'datatype', 'default']},
  }

  const controller = new AbortController()

  const request = fetch(url, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
    signal: controller.signal,
  })

  const promise = request
    .then(processResponse)
    .catch(e =>
      e.name === 'AbortError'
        ? Promise.reject(new CancellationError())
        : Promise.reject(e)
    )

  return {
    promise,
    cancel: () => controller.abort(),
  }
}

const processResponse = async (response: Response): Promise<RunQueryResult> => {
  const reader = response.body.getReader()
  const decoder = new TextDecoder()

  let csv = ''
  let bytesRead = 0
  let didTruncate = false

  let read = await reader.read()

  while (!read.done) {
    const text = decoder.decode(read.value)

    bytesRead += read.value.byteLength

    if (bytesRead > FLUX_RESPONSE_BYTES_LIMIT) {
      csv += trimPartialLines(text)
      didTruncate = true
      break
    } else {
      csv += text
      read = await reader.read()
    }
  }

  reader.cancel()

  return {
    csv,
    bytesRead,
    didTruncate,
  }
}

/*
  Given an arbitrary text chunk of a Flux CSV, trim partial lines off of the end
  of the text.

  For example, given the following partial Flux response,

            r,baz,3
      foo,bar,baz,2
      foo,bar,b

  we want to trim the last incomplete line, so that the result is

            r,baz,3
      foo,bar,baz,2

*/
const trimPartialLines = (partialResp: string): string => {
  let i = partialResp.length - 1

  while (partialResp[i] !== '\n') {
    if (i <= 0) {
      return partialResp
    }

    i -= 1
  }

  return partialResp.slice(0, i + 1)
}