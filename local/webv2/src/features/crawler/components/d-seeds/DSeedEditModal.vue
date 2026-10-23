<template>
  <div>
    <div
      id="dSeedEditModal"
      class="modal fade"
      :class="{ 'show d-block': active }"
      tabindex="-1"
      :aria-hidden="active ? 'false' : 'true'"
    >
      <div class="modal-dialog modal-dialog-centered modal-lg">
        <div class="modal-content">
          <form id="dSeedEditForm" @submit.prevent="$emit('submit')">
            <div class="modal-header">
              <div>
                <h5 id="dSeedEditModalTitle" class="modal-title">
                  {{ model.id ? "编辑 d_seed 条目" : "新增 d_seed 条目" }}
                </h5>
                <div id="dSeedEditModalMeta" class="small text-muted">
                  {{ model.id ? model.name : "创建新的查询配置" }}
                </div>
              </div>
              <button type="button" class="btn-close" aria-label="Close" @click="$emit('close')"></button>
            </div>

            <div class="modal-body">
              <input id="dSeedEditId" type="hidden" :value="model.id" />
              <div class="row g-3">
                <div class="col-12 col-md-6">
                  <label class="form-label mb-1" for="dSeedEditName">查询名</label>
                  <input
                    id="dSeedEditName"
                    class="form-control"
                    type="text"
                    maxlength="128"
                    :value="model.name"
                    @input="set('name', inputValue($event))"
                  />
                </div>
                <div class="col-12 col-md-3">
                  <label class="form-label mb-1" for="dSeedEditActive">启用状态</label>
                  <select
                    id="dSeedEditActive"
                    class="form-select"
                    :value="model.active"
                    @change="set('active', selectValue($event))"
                  >
                    <option value="2">启用</option>
                    <option value="1">停用</option>
                  </select>
                </div>
                <div class="col-12 col-md-3">
                  <label class="form-label mb-1" for="dSeedEditLastStatus">最后状态</label>
                  <select
                    id="dSeedEditLastStatus"
                    class="form-select"
                    :value="model.lastStatus"
                    @change="set('lastStatus', selectValue($event))"
                  >
                    <option value="0">空闲</option>
                    <option value="1">成功</option>
                    <option value="2">空页</option>
                    <option value="3">错误</option>
                  </select>
                </div>
                <div class="col-12 col-md-3">
                  <label class="form-label mb-1" for="dSeedEditNameType">名称类型</label>
                  <select
                    id="dSeedEditNameType"
                    class="form-select"
                    :value="model.nameType"
                    @change="set('nameType', selectValue($event))"
                  >
                    <option value="1">Prefix</option>
                    <option value="2">Label</option>
                  </select>
                </div>
                <div class="col-12 col-md-3">
                  <label class="form-label mb-1" for="dSeedEditSearchType">查询类型</label>
                  <select
                    id="dSeedEditSearchType"
                    class="form-select"
                    :value="model.searchType"
                    @change="set('searchType', selectValue($event))"
                  >
                    <option value="1">偏移量</option>
                    <option value="2">起止页</option>
                  </select>
                </div>
                <NumberField
                  id="dSeedEditPageNow"
                  label="当前页"
                  field="pageNow"
                  :value="model.pageNow"
                  @update="set"
                />
                <NumberField
                  v-if="model.searchType === '1'"
                  id="dSeedEditOffset"
                  label="偏移量"
                  field="offset"
                  :value="model.offset"
                  wrapper-id="dSeedOffsetFieldWrap"
                  @update="set"
                />
                <template v-else>
                  <NumberField
                    id="dSeedEditStartPage"
                    label="起始页"
                    field="startPage"
                    :value="model.startPage"
                    wrapper-id="dSeedStartPageFieldWrap"
                    @update="set"
                  />
                  <NumberField
                    id="dSeedEditEndPage"
                    label="结束页"
                    field="endPage"
                    :value="model.endPage"
                    wrapper-id="dSeedEndPageFieldWrap"
                    @update="set"
                  />
                </template>
                <div class="col-12">
                  <label class="form-label mb-1" for="dSeedEditLastError">最后错误</label>
                  <textarea
                    id="dSeedEditLastError"
                    class="form-control"
                    rows="3"
                    :value="model.lastError"
                    @input="set('lastError', inputValue($event))"
                  ></textarea>
                </div>
              </div>
              <div id="dSeedEditMsg" class="alert d-none mt-3 mb-0" role="alert"></div>
            </div>

            <div class="modal-footer">
              <button type="button" class="btn btn-light" @click="$emit('close')">取消</button>
              <button id="dSeedEditSubmit" type="submit" class="btn btn-primary" :disabled="submitting">
                {{ submitting ? "保存中..." : "保存" }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
    <div v-if="active" class="modal-backdrop fade show" @click="$emit('close')"></div>
  </div>
</template>

<script setup lang="ts">
import NumberField from "./DSeedNumberField.vue";
import type { DSeedEditModel } from "./models";

defineProps<{
  active: boolean;
  model: DSeedEditModel;
  submitting: boolean;
}>();

const emit = defineEmits<{
  close: [];
  update: [patch: Partial<DSeedEditModel>];
  submit: [];
}>();

function set(key: keyof DSeedEditModel, value: string) {
  emit("update", { [key]: value });
}

function inputValue(event: Event): string {
  return (event.target as HTMLInputElement | HTMLTextAreaElement).value;
}

function selectValue(event: Event): string {
  return (event.target as HTMLSelectElement).value;
}
</script>
