export interface Package {
  name: string
  description?: string
  version?: string
  private?: boolean
  updatedAt?: string
}

export interface PackageMetadata {
  name: string
  description?: string
  'dist-tags': Record<string, string>
  versions: Record<string, PackageVersion>
  time: Record<string, string>
  readme?: string
}

export interface PackageVersion {
  name: string
  version: string
  description?: string
  main?: string
  license?: string
  homepage?: string
  repository?: {
    type?: string
    url?: string
  }
  dependencies?: Record<string, string>
  devDependencies?: Record<string, string>
  dist?: {
    tarball: string
    shasum: string
    size?: number
  }
}

export interface User {
  username: string
  email: string
  role: string
  lastLogin?: string
  createdAt?: string
}
