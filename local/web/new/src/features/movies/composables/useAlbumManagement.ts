import { ref } from "vue";

export function useAlbumManagement<T extends string>() {
  const activeModal = ref<T | "">("");
  const open = (modal: T): void => {
    activeModal.value = modal;
  };
  const close = (): void => {
    activeModal.value = "";
  };
  return { activeModal, open, close };
}
