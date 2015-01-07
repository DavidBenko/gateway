import Ember from "ember";

export var ProxyEndpointsController = Ember.ArrayController.extend({
  sortProperties: ['name'],
  sortAscending: true
});

export var ProxyEndpointController = Ember.ObjectController.extend({
  needs: ['admin'],

  actions: {
    save: function() {
      var self = this;
      this.model.save().then(function() {
        self.set('controllers.admin.successMessage', "Saved!");
        self.set('controllers.admin.errorMessage', null);
      }, function(reason) {
        self.set('controllers.admin.successMessage', null);
        self.set('controllers.admin.errorMessage', reason.responseText);
      });
    },
    delete: function() {
      if (confirm("Delete the endpoint '" + this.model.get('name') + "'?")) {
        this.model.destroyRecord();
        this.set('controllers.admin.successMessage', "Deleted!");
        this.set('controllers.admin.errorMessage', null);
        this.transitionToRoute('endpoints');
      }
    }
  }
});

export var NewProxyEndpointController = Ember.ObjectController.extend({
  // This is almost entirely duplicated from EndpointController,
  // but specifying controllerName in my route wouldn't resolve the
  // 'save' action.

  needs: ['admin'],

  actions: {
    save: function() {
      var self = this;
      this.model.save().then(function(value) {
        self.set('controllers.admin.successMessage', "Created!");
        self.set('controllers.admin.errorMessage', null);
        self.transitionToRoute("endpoint", value.id)
      }, function(reason) {
        self.set('controllers.admin.successMessage', null);
        self.set('controllers.admin.errorMessage', reason.responseText);
      });
    }
  }
});


export default ProxyEndpointsController;