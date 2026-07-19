import { ref } from "vue";

export function useMovieAlbums() {
  const showMovieAlbumModal = ref(false);
  const selectedMovieAlbumName = ref("");
  const newMovieAlbumName = ref("");
  const movieAlbumBusy = ref(false);
  const movieAlbumMessage = ref("");
  const movieAlbumMessageOk = ref(true);

  function reset(): void {
    showMovieAlbumModal.value = false;
    newMovieAlbumName.value = "";
    movieAlbumMessage.value = "";
    movieAlbumMessageOk.value = true;
  }

  return {
    showMovieAlbumModal,
    selectedMovieAlbumName,
    newMovieAlbumName,
    movieAlbumBusy,
    movieAlbumMessage,
    movieAlbumMessageOk,
    reset
  };
}
