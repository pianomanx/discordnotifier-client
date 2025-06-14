import { get } from 'svelte/store'
import { getUi, type BackendResponse } from '../../api/fetch'
import { type TriggerInfo } from '../../api/notifiarrConfig'
import { profile } from '../../api/profile.svelte'
import { reload } from '../../header/Reload.svelte'
import { warning } from '../../includes/util'

const reloadClient = async () => {
  try {
    await reload()
    await profile.refresh()
  } catch (error) {
    warning(`${error}`)
  }
}

// Some triggers require a specific input to be passed. This function returns that input value.
export const option = (info: TriggerInfo): any => {
  if (info.key == 'TrigCustomCommand')
    // This one requires a hash of the command.
    return get(profile).config.commands?.find(c => c.name == info.name)?.hash

  if (info.key == 'TrigCustomCronTimer')
    // This one requires the index of the cron.
    get(profile).siteCrons?.forEach((cron, i) => {
      if (info.name.endsWith("'" + cron.name + "'")) return i
    })

  // This one requires the name of the endpoint.
  if (info.key == 'TrigEndpointURL') return info.name

  return ''
}

export const run = async (info: TriggerInfo, content?: any): Promise<BackendResponse> => {
  const now = Date.now()
  if (info.key == 'TrigStop') {
    // Special case for the reload client trigger.
    await reloadClient()
    return { ok: true, body: '' }
  }

  const url = ['trigger', info.key, encodeURIComponent(content ?? option(info))]
    .filter(v => v)
    .join('/')
  const resp = await getUi(url + '?ts=' + now, false)
  if (resp.ok) await profile.refresh()
  else return resp
  const diff = Date.now() - now
  // It always takes at least 1 second to run.
  if (diff < 1000) await new Promise(resolve => setTimeout(resolve, 1000 - diff))
  return resp
}
