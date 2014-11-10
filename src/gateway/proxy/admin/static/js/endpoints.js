App.EndpointsIndexRoute = Ember.Route.extend({
  beforeModel: function() {
    this.transitionTo('newEndpoint');
  }
});

App.Endpoint = DS.Model.extend({
  name: DS.attr(),
  script: DS.attr()
});

App.EndpointsRoute = Ember.Route.extend({
  model: function() {
    return this.store.find('endpoint');
  }
});

App.EndpointsController = Ember.ArrayController.extend({
  sortProperties: ['name'],
  sortAscending: true
})

App.EndpointRoute = Ember.Route.extend({
  model: function(params) {
    return this.store.find('endpoint', params.endpoint_id);
  }
})

App.EndpointController = Ember.ObjectController.extend({
  needs: ['admin'],

  actions: {
    save: function() {
      var self = this;
      this.model.save().then(function(value) {
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

App.NewEndpointRoute = Ember.Route.extend({
  templateName: 'endpoint',
  model: function(params) {
    return this.store.createRecord('endpoint');
  }
})

App.NewEndpointController = Ember.ObjectController.extend({
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
