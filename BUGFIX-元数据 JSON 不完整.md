# Bug 修复：包元数据 JSON 不完整问题

## 🐛 问题描述

### 错误现象

```bash
pnpm i
WARN  GET http://localhost:4873/vite error (undefined). Will retry in 10 seconds.
ERR_PNPM_BROKEN_METADATA_JSON  Unterminated string in JSON at position 5241983
```

### 错误分析

**错误位置**: `position 5241983` (约 5MB 处)

**根本原因**:
1. Grape 从上游 (npmjs) 获取包元数据时，某些大型包（如 vite、typescript）的元数据非常大
2. 在保存或返回时，**JSON 数据被截断**，导致不完整
3. pnpm 解析时遇到无效的 JSON 格式

---

## ✅ 修复方案

### 1. 增加 JSON 完整性验证

**文件**: `internal/registry/proxy.go`

**修改**:
```go
// GetMetadata 从上游获取包元数据
func (p *Proxy) GetMetadata(packageName string) ([]byte, error) {
    // ... 原有代码 ...
    
    // 新增：验证 JSON 完整性
    if err := validateJSON(data); err != nil {
        logger.Warnf("Invalid JSON from upstream [%s] for package %s: %v", up.Name, packageName, err)
        return nil, fmt.Errorf("invalid JSON from upstream: %w", err)
    }
    
    return data, nil
}

// validateJSON 验证 JSON 是否完整有效
func validateJSON(data []byte) error {
    var raw json.RawMessage
    return json.Unmarshal(data, &raw)
}
```

---

### 2. 存储层增加验证和原子写入

**文件**: `internal/storage/local/storage.go`

**修改 1 - GetMetadata 增加验证**:
```go
func (s *Storage) GetMetadata(packageName string) ([]byte, error) {
    // ... 原有代码 ...
    
    // 新增：验证 JSON 完整性
    if err := validateMetadataJSON(data); err != nil {
        logger.Warnf("Corrupted metadata for package %s: %v", packageName, err)
        // 如果数据损坏，删除并返回不存在，让上层重新获取
        os.Remove(path)
        return nil, registry.ErrPackageNotFound
    }
    
    return data, nil
}

// validateMetadataJSON 验证元数据 JSON 是否完整
func validateMetadataJSON(data []byte) error {
    var raw json.RawMessage
    return json.Unmarshal(data, &raw)
}
```

**修改 2 - SaveMetadata 使用原子写入**:
```go
func (s *Storage) SaveMetadata(packageName string, data []byte) error {
    // ... 原有代码 ...
    
    // 新增：验证 JSON 完整性
    if err := validateMetadataJSON(data); err != nil {
        return fmt.Errorf("invalid metadata JSON: %w", err)
    }
    
    // 原子写入：先写入临时文件，再重命名
    tmpPath := path + ".tmp"
    if err := os.WriteFile(tmpPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write metadata: %w", err)
    }
    
    // 重命名临时文件到目标文件（原子操作）
    if err := os.Rename(tmpPath, path); err != nil {
        os.Remove(tmpPath) // 清理临时文件
        return fmt.Errorf("failed to finalize metadata: %w", err)
    }
    
    return nil
}
```

---

### 3. Handler 层增加错误处理

**文件**: `internal/server/handler/registry.go`

**修改**:
```go
func (h *RegistryHandler) rewriteTarballURLs(data []byte, packageName string, baseURL string) ([]byte, error) {
    // ... 原有代码 ...
    
    pkg["versions"] = versions
    
    // 使用 json.Marshal 确保输出有效的 JSON
    rewritten, err := json.Marshal(pkg)
    if err != nil {
        logger.Errorf("Failed to marshal rewritten JSON for %s: %v", packageName, err)
        return data, nil // 返回原始数据
    }
    
    return rewritten, nil
}
```

---

## 🔧 额外修复

### 清理损坏的缓存

```bash
# 清理可能已损坏的包缓存
rm -rf ./data/packages/vite
rm -rf ./data/packages/typescript
```

---

## ✅ 验证测试

### 1. 重新编译

```bash
make build-only
```

### 2. 启动服务

```bash
./bin/grape
```

### 3. 测试安装大型包

```bash
cd test-projects/vue3-demo

# 清理 node_modules
rm -rf node_modules package-lock.json

# 重新安装
npm install
```

### 4. 预期结果

```bash
✅ 安装成功，无 JSON 解析错误
✅ pnpm install 正常完成
✅ yarn install 正常完成
```

---

## 📊 修复效果

| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| 安装小型包 | ✅ 正常 | ✅ 正常 |
| 安装大型包 | ❌ JSON 截断 | ✅ 完整验证 |
| 缓存损坏 | ❌ 持续报错 | ✅ 自动恢复 |
| 并发写入 | ❌ 可能损坏 | ✅ 原子写入 |

---

## 🛡️ 防护机制

现在 Grape 有**三层防护**确保 JSON 完整性：

1. **上游获取时验证** - 从 npmjs 获取后立即验证
2. **保存前验证** - 保存到磁盘前再次验证
3. **读取时验证** - 从磁盘读取后验证，发现损坏自动删除

---

## 📝 技术要点

### 1. JSON 验证

使用 `json.RawMessage` 延迟解析，仅验证格式：

```go
func validateJSON(data []byte) error {
    var raw json.RawMessage
    return json.Unmarshal(data, &raw)
}
```

### 2. 原子写入

使用临时文件 + 重命名确保原子性：

```go
tmpPath := path + ".tmp"
os.WriteFile(tmpPath, data, 0644)
os.Rename(tmpPath, path)  // 原子操作
```

### 3. 自动恢复

发现损坏数据时自动删除，触发重新获取：

```go
if err := validateMetadataJSON(data); err != nil {
    os.Remove(path)  // 删除损坏文件
    return nil, registry.ErrPackageNotFound  // 触发重新获取
}
```

---

## 🚀 后续优化建议

1. **增加压缩存储** - 大型包元数据压缩后保存
2. **增量更新** - 仅更新变化的版本信息
3. **缓存过期策略** - 定期清理过期缓存
4. **并发控制** - 防止同一包并发写入冲突

---

## 📚 相关文件

- `internal/registry/proxy.go` - 上游代理层
- `internal/storage/local/storage.go` - 存储层
- `internal/server/handler/registry.go` - Handler 层

---

**修复完成时间**: 2026-02-26  
**影响范围**: 所有包元数据缓存  
**向后兼容**: ✅ 完全兼容
