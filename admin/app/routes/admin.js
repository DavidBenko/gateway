import Ember from "ember";

var AdminIndexRoute = Ember.Route.extend({
  beforeModel: function() {
    this.transitionTo('proxyEndpoints');
  }
});

export default AdminIndexRoute;