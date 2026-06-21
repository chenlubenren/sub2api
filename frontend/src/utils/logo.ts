const logoCache = new Map<string, Promise<string>>()

function isProcessablePng(src: string) {
  return /^data:image\/png/i.test(src) || /\.png(?:$|\?)/i.test(src)
}

function loadImage(src: string) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const image = new Image()
    image.crossOrigin = 'anonymous'
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('Failed to load logo image'))
    image.src = src
  })
}

export function prepareLogoAsset(src: string): Promise<string> {
  if (!src || !isProcessablePng(src)) {
    return Promise.resolve(src)
  }

  const cached = logoCache.get(src)
  if (cached) {
    return cached
  }

  const task = loadImage(src)
    .then((image) => {
      const canvas = document.createElement('canvas')
      canvas.width = image.naturalWidth || image.width
      canvas.height = image.naturalHeight || image.height

      const context = canvas.getContext('2d', { willReadFrequently: true })
      if (!context) {
        return src
      }

      context.drawImage(image, 0, 0)
      const frame = context.getImageData(0, 0, canvas.width, canvas.height)
      const { data, width, height } = frame

      let minX = width
      let minY = height
      let maxX = -1
      let maxY = -1

      for (let y = 0; y < height; y += 1) {
        for (let x = 0; x < width; x += 1) {
          const index = (y * width + x) * 4
          const r = data[index]
          const g = data[index + 1]
          const b = data[index + 2]
          const alpha = data[index + 3]

          if (alpha === 0) {
            continue
          }

          const nearWhite = r > 245 && g > 245 && b > 245
          if (nearWhite) {
            data[index + 3] = 0
            continue
          }

          if (x < minX) minX = x
          if (y < minY) minY = y
          if (x > maxX) maxX = x
          if (y > maxY) maxY = y
        }
      }

      context.putImageData(frame, 0, 0)

      if (maxX < minX || maxY < minY) {
        return src
      }

      const cropWidth = maxX - minX + 1
      const cropHeight = maxY - minY + 1
      const outputSize = Math.max(cropWidth, cropHeight)
      const outputCanvas = document.createElement('canvas')
      outputCanvas.width = outputSize
      outputCanvas.height = outputSize

      const outputContext = outputCanvas.getContext('2d')
      if (!outputContext) {
        return src
      }

      const offsetX = Math.floor((outputSize - cropWidth) / 2)
      const offsetY = Math.floor((outputSize - cropHeight) / 2)

      outputContext.drawImage(
        canvas,
        minX,
        minY,
        cropWidth,
        cropHeight,
        offsetX,
        offsetY,
        cropWidth,
        cropHeight
      )

      return outputCanvas.toDataURL('image/png')
    })
    .catch(() => src)

  logoCache.set(src, task)
  return task
}
