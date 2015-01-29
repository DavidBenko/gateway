require_relative './spec_helper'

describe "users via accounts" do

  describe "index" do
    context "empty" do
      it "should return 200"
      it "should return an array in 'users'"
    end

    context "with data" do
      it "should return all users in account"
    end

    context "non-existent account" do
      it "should return 404"
    end
  end

  describe "create" do
    context "with valid data" do
      it "should return 200"
      it "should return the new ID"
      it "should return the name"
      it "should return the email"
      it "should not return the password"
      it "should not return the hashed_password"
    end

    context "with invalid json" do
      it "should return 400"
      it "should return an error"
    end

    context "without a name" do
      it "should return 400"
      it "should return an error"
    end

    context "with a duplicate name" do
      it "should return 400"
      it "should return an error"
    end

    context "without an email" do
      it "should return 400"
      it "should return an error"
    end

    context "with a duplicate email" do
      it "should return 400"
      it "should return an error"
    end

    context "without a password" do
      it "should return 400"
      it "should return an error"
    end

    context "without a matching password confirmation" do
      it "should return 400"
      it "should return an error"
    end
  end

  describe "show" do
    context "existing" do
      it "should return 200"
      it "should return the id"
      it "should return the name"
      it "should return the email"
      it "should not return the password"
      it "should not return the hashed_password"
    end

    context "non-existing" do
      it "should return 404"
    end

    context "mismatched account" do
      it "should return 404"
    end
  end

  describe "update" do
    context "with valid data" do
      it "should return 200"
      it "should return the same ID"
      it "should return the new name"
    end

    context "without updating password" do
      it "should return 200"
      it "should not change the password"
    end

    context "updating password" do
      it "should return 200"
      it "should change the password"
    end


    context "with invalid json" do
      it "should return 400"
      it "should return an error"
    end

    context "without a name" do
      it "should return 400"
      it "should return an error"
    end

    context "with a duplicate name" do
      it "should return 400"
      it "should return an error"
    end

    context "without an email" do
      it "should return 400"
      it "should return an error"
    end

    context "with a duplicate email" do
      it "should return 400"
      it "should return an error"
    end

    context "with a password but without a matching password confirmation" do
      it "should return 400"
      it "should return an error"
    end

    context "non-existing" do
      it "should return 404"
    end

    context "mismatched account" do
      it "should return 404"
    end
  end

  describe "delete" do
    context "existing" do
      it "should return 200"
      it "should return nothing"
      it "should remove the item"
    end

    context "non-existing" do
      it "should return 404"
    end

    context "mismatched account" do
      it "should return 404"
    end
  end

end
