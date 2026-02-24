<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { login } from '@/api/auth'
import { ElMessage } from 'element-plus'
import { Truck, Lock, User as UserIcon, Loader2 } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const handleLogin = async () => {
  if (!loginForm.username || !loginForm.password) {
    ElMessage.warning('Please enter username and password')
    return
  }
  
  loading.value = true
  try {
    const res: any = await login(loginForm)
    authStore.setAuth(res.token, res.user)
    ElMessage.success('Login successful')
    router.push('/dashboard')
  } catch (err) {
    // Error handled by interceptor
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <div class="background-overlay"></div>
    
    <div class="login-card glass-panel">
      <div class="login-header">
        <div class="logo-circle">
          <Truck class="logo-icon" />
        </div>
        <h1>Rescue EMS</h1>
        <p>Emergency Management System</p>
      </div>

      <el-form :model="loginForm" class="login-form">
        <el-form-item>
          <el-input
            v-model="loginForm.username"
            placeholder="Username"
            :prefix-icon="UserIcon"
            size="large"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="Password"
            :prefix-icon="Lock"
            size="large"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        
        <el-button
          type="primary"
          class="login-button"
          :loading="loading"
          size="large"
          @click="handleLogin"
        >
          {{ loading ? 'Authenticating...' : 'Sign In' }}
        </el-button>
      </el-form>

      <div class="login-footer">
        <span>Forgot password? Contact Administrator</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  height: 100vh;
  width: 100vw;
  display: flex;
  align-items: center;
  justify-content: center;
  background-image: url('https://images.unsplash.com/photo-1582213713364-2ba1555f2cd5?auto=format&fit=crop&q=80&w=2070');
  background-size: cover;
  background-position: center;
  position: relative;
  overflow: hidden;
}

.background-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: radial-gradient(circle at center, rgba(13, 17, 23, 0.4) 0%, rgba(13, 17, 23, 0.9) 100%);
}

.login-card {
  width: 100%;
  max-width: 420px;
  padding: 40px;
  border-radius: 20px;
  position: relative;
  z-index: 10;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  background: rgba(13, 17, 23, 0.7);
  backdrop-filter: blur(20px);
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.logo-circle {
  width: 64px;
  height: 64px;
  background: rgba(59, 130, 246, 0.1);
  border: 2px solid #3b82f6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.4);
}

.logo-icon {
  width: 32px;
  height: 32px;
  color: #3b82f6;
}

.login-header h1 {
  font-size: 2rem;
  margin: 0;
  color: #fff;
  letter-spacing: 1px;
}

.login-header p {
  color: #8b949e;
  margin-top: 8px;
}

.login-form {
  margin-bottom: 24px;
}

.login-button {
  width: 100%;
  height: 50px;
  font-size: 1.1rem;
  font-weight: 600;
  border-radius: 12px;
  margin-top: 10px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border: none;
}

.login-footer {
  text-align: center;
  color: #6e7681;
  font-size: 0.9rem;
}

/* Override Element Plus styles */
:deep(.el-input__wrapper) {
  background-color: rgba(48, 54, 61, 0.5) !important;
  box-shadow: none !important;
  border: 1px solid #30363d !important;
  border-radius: 12px;
}

:deep(.el-input__inner) {
  color: #fff !important;
}

:deep(.el-input__inner::placeholder) {
  color: #484f58;
}
</style>
