<script setup lang="ts">
import {reactive, ref, computed, watch} from 'vue'
import {Run} from '../../wailsjs/go/main/App'
import {EventsOn} from '../../wailsjs/runtime'
import {Message} from "@arco-design/web-vue";

enum modeEnum {
  GATHER = 'gather',
  EXEC = 'exec',
}

type modeEnumListType = {
  key: '',
  value: ''
}

const modeEnumList = ref<modeEnumListType[]>([] as modeEnumListType[])
const configFileList = ref<string[]>([])

const formData = reactive({
  mode: modeEnum.GATHER,
  saveConfigName: '',
  runConfigName: '',
  minInterval: 0,
  cycle: 0

})
const exec = async () => {
  let configName = formData.mode === modeEnum.GATHER ? formData.saveConfigName : formData.runConfigName
  if (configName === '') {
    Message.error({
      content: formData.mode === modeEnum.GATHER ? '请输入配置文件名' : '请选择需要执行的配置',
      id: 'message'
    })
    return;
  }
  const res = await Run({
    mode: formData.mode,
    configName,
    minInterval: formData.minInterval,
    cycle: formData.cycle
  })
  if (res && formData.mode === modeEnum.GATHER && !configFileList.value.includes(configName)) {
    configFileList.value.unshift(configName)
  }
}

const updateSelectDropdownMaxHeight = () => {
  const windowHeight = window.innerHeight
  const maxHeight = 200
  const minHeight = 60
  let selectDropdownMaxHeight = Math.floor(windowHeight / 2.5)

  if (selectDropdownMaxHeight < minHeight) {
    selectDropdownMaxHeight = minHeight
  } else if (selectDropdownMaxHeight > maxHeight) {
    selectDropdownMaxHeight = maxHeight
  }

  const selectContent = document.querySelector('.arco-select-dropdown-list-wrapper')
  if (selectContent) {
    selectContent.setAttribute('style', `max-height:${selectDropdownMaxHeight}px`)
  }
}

EventsOn('init', (res) => {
  formData.saveConfigName = res.configName
  formData.minInterval = res.minInterval
  configFileList.value = res.configFileList
  modeEnumList.value = res.modeEnumList
  updateSelectDropdownMaxHeight()
})


//监听窗口变化
window.addEventListener('resize', () => {
  updateSelectDropdownMaxHeight()
})


const alertConfig = reactive({
  show: false,
  msg: '',
  isMainNotAlert: true,
  type: 'info' as 'info' | 'success' | 'warning' | 'error',
  close: () => {
    alertConfig.show = false
    alertConfig.isMainNotAlert = true
  }
})


let isReceiveMessages = true
EventsOn('alertMsg', (msg) => {
  if (!isReceiveMessages) {
    return
  }
  alertConfig.show = true
  alertConfig.msg = msg
  alertConfig.isMainNotAlert = false
  alertConfig.type = 'info'
  if (msg === '程序已退出运行') {
    alertConfig.type = 'error'
    isReceiveMessages = false
    setTimeout(() => {
      isReceiveMessages = true
    }, 500)
  }
})

const gatherFormItem = computed(() => {
  return formData.mode === modeEnum.GATHER
})

watch(
    () => formData.saveConfigName,
    (newVal) => {
      formData.runConfigName = newVal
    }
)

</script>
<template>
  <div class="main" :class="{'main-not-alert': alertConfig.isMainNotAlert}">
    <a-alert :type="alertConfig.type" v-if="alertConfig.show" :closable="true" @afterClose="alertConfig.close">
      {{ alertConfig.msg }}
    </a-alert>
    <a-form :model="formData" :style="{ width: '320px' }" :label-col-props="{span:8}"
            :wrapper-col-props="{span:16}">
      <a-form-item label="运行模式">
        <a-radio-group v-model="formData.mode">
          <a-radio v-for="item in modeEnumList" :key="item.key" :value="item.key">{{ item.value }}</a-radio>
        </a-radio-group>
      </a-form-item>
      <div v-if="gatherFormItem">
        <a-form-item label="配置文件名"
                     :row-props="{ justify: 'start' }">
          <a-input v-model="formData.saveConfigName"></a-input>
        </a-form-item>
        <a-form-item label="最小间隔" extra="每次点击之间的最小时间间隔">
          <a-input-number mode="button" :min="0" v-model="formData.minInterval">
            <template #suffix>
              <span>毫秒</span>
            </template>
          </a-input-number>
        </a-form-item>
      </div>
      <div v-else>
        <a-form-item label="执行配置">
          <a-select v-model="formData.runConfigName" placeholder="请选择需要执行的配置">
            <a-option v-for="item in configFileList" :key="item" :value="item">{{ item }}</a-option>
          </a-select>
        </a-form-item>
        <a-form-item label="执行周期" extra="0次代表无限循环">
          <a-input-number mode="button" :min="0" v-model="formData.cycle">
            <template #suffix>
              <span>次</span>
            </template>
          </a-input-number>
        </a-form-item>
      </div>
      <a-form-item>
        <a-button type="primary" @click="exec">运行</a-button>
      </a-form-item>
    </a-form>
  </div>
</template>

<style scoped>
.main-not-alert {
  padding-top: 10vh;
}
</style>
