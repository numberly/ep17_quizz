<template>
  <div id="app">
    <!-- <img src="./assets/logo.png"> -->
    <router-view></router-view>
  </div>
</template>

<script>
export default {
  name: 'app',
  beforeCreate () {
    this.$store.dispatch('getQuiz')
  }
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

html, body, #app {
  padding:0;
  margin:0;
  height: 100%;
}
body {
  background-image: url('assets/img/bgslide-1.svg');
  background-size: cover;
  background-position: 50%;
  background: -webkit-linear-gradient(to right, #7b4397, #dc2430);
  background: linear-gradient(to right, #7b4397, #dc2430);
}

h1,h2,h3,h4,h5,h6 {
  color: white;
}

/*.el-input__inner,
.el-input__inner:focus,
.el-input__inner:hover {
  color: white;
  background: rgba(255, 255, 255, 0.2);
  border-color: transparent;
}*/

/*.el-form-item__label {
  color: white;
}*/
.el-autocomplete-suggestion li {
  color: #444;
}

/* DIRTY */

@media all and (orientation:portrait) {
  .el-form-item__label {
    font-size: 18px;
  }
  .el-form-item__content {
    font-size: 16px;
  }
  .el-input {
    font-size: 18px;
  }
  .el-button {
    font-size: 18px;
  }
}
</style>
