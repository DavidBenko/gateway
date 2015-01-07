import Ember from "ember";

var AdminController = Ember.ObjectController.extend({
  errorMessage: null,
  successMessage: null,

  actions: {
    closeSuccess: function() {
      this.set('successMessage', null);
    },
    closeError: function() {
      this.set('errorMessage', null);
    }
  }
});

export default AdminController;
