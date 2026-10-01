import { apiClient } from '../client'

export type ModelPricingPlatform = 'openai' | 'anthropic'

/** Standard price card values are USD per token. The UI displays $ / 1M. */
export interface GlobalModelPricing {
  model: string
  label: string
  platform: ModelPricingPlatform
  input_price: number
  output_price: number
  cache_write_price: number
  cache_read_price: number
}

interface GlobalModelPricingResponse {
  models: GlobalModelPricing[]
}

export async function getGlobalModelPricing(): Promise<GlobalModelPricing[]> {
  const { data } = await apiClient.get<GlobalModelPricingResponse>('/admin/model-pricing')
  return data.models
}

export async function updateGlobalModelPricing(models: GlobalModelPricing[]): Promise<GlobalModelPricing[]> {
  const { data } = await apiClient.put<GlobalModelPricingResponse>('/admin/model-pricing', { models })
  return data.models
}

export default { getGlobalModelPricing, updateGlobalModelPricing }
